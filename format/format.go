package format

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mreimbold/terraformat/config"
)

type itemKind int

const (
	itemAttribute itemKind = iota
	itemBlock
)

type bodyContext struct {
	root      bool
	blockType string
}

type bodyItem struct {
	kind      itemKind
	attr      *hclwrite.Attribute
	block     *hclwrite.Block
	name      string
	labelKey  string
	tokens    hclwrite.Tokens
	prefix    hclwrite.Tokens
	origIndex int
	start     int
	end       int
}

// Format applies terraformat rules to a Terraform/HCL document.
func Format(src []byte, cfg config.Config) ([]byte, error) {
	file, diags := hclwrite.ParseConfig(src, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse: %s", diags.Error())
	}

	if err := rewriteBody(file.Body(), bodyContext{root: true}, cfg); err != nil {
		return nil, err
	}

	out := file.Bytes()
	if cfg.EnsureEOFNewline {
		out = ensureTrailingNewline(out)
	}
	return out, nil
}

func rewriteBody(body *hclwrite.Body, ctx bodyContext, cfg config.Config) error {
	// Rewrite nested blocks first to avoid losing structure when we flatten bodies.
	for _, block := range body.Blocks() {
		if err := rewriteBody(block.Body(), bodyContext{blockType: block.Type()}, cfg); err != nil {
			return err
		}
	}

	leading, items, trailing, err := collectBodyItems(body)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	applyOrdering := cfg.EnforceAttributeOrder || (cfg.EnforceBlockOrder && ctx.root)
	if applyOrdering {
		sortBodyItems(items, ctx, cfg)
	}

	newTokens := renderBody(leading, items, trailing, ctx, cfg)
	body.Clear()
	body.AppendUnstructuredTokens(newTokens)
	return nil
}

func collectBodyItems(body *hclwrite.Body) (hclwrite.Tokens, []bodyItem, hclwrite.Tokens, error) {
	attrs := body.Attributes()
	attrNames := make(map[*hclwrite.Attribute]string, len(attrs))
	for name, attr := range attrs {
		attrNames[attr] = name
	}

	items := make([]bodyItem, 0, len(attrs)+len(body.Blocks()))
	for _, attr := range attrs {
		items = append(items, bodyItem{
			kind:   itemAttribute,
			attr:   attr,
			name:   attrNames[attr],
			tokens: attr.BuildTokens(nil),
		})
	}
	for _, block := range body.Blocks() {
		items = append(items, bodyItem{
			kind:     itemBlock,
			block:    block,
			name:     block.Type(),
			labelKey: strings.Join(block.Labels(), "."),
			tokens:   block.BuildTokens(nil),
		})
	}

	if len(items) == 0 {
		return nil, nil, nil, nil
	}

	bodyTokens := body.BuildTokens(nil)
	for i := range items {
		start, end, ok := findTokenSpan(bodyTokens, items[i].tokens)
		if !ok {
			return nil, nil, nil, fmt.Errorf("failed to locate item tokens in body")
		}
		items[i].start = start
		items[i].end = end
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].start < items[j].start
	})
	for i := range items {
		items[i].origIndex = i
	}

	prevEnd := -1
	for i := range items {
		items[i].prefix = bodyTokens[prevEnd+1 : items[i].start]
		prevEnd = items[i].end
	}
	trailing := bodyTokens[prevEnd+1:]

	leading := items[0].prefix
	items[0].prefix = nil

	return leading, items, trailing, nil
}

func renderBody(leading hclwrite.Tokens, items []bodyItem, trailing hclwrite.Tokens, ctx bodyContext, cfg config.Config) hclwrite.Tokens {
	out := normalizeLeadingTokens(leading)
	for i, item := range items {
		insertBlank := i > 0 && cfg.EnforceTopLevelSpacing && ctx.root && items[i-1].kind == itemBlock && item.kind == itemBlock
		if insertBlank {
			out = append(out, newlineToken())
		}
		prefix := normalizePrefixTokens(item.prefix)
		if insertBlank && !containsComment(item.prefix) {
			prefix = nil
		}
		out = append(out, prefix...)
		out = append(out, item.tokens...)
	}
	out = append(out, trailing...)
	return out
}

func ensureTrailingNewline(src []byte) []byte {
	if len(src) == 0 || bytes.HasSuffix(src, []byte("\n")) {
		return src
	}
	return append(src, '\n')
}
