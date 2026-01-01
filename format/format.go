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

type staticError string

// Error returns the error string.
func (err staticError) Error() string {
	return string(err)
}

const (
	errParseConfig    staticError = "parse config"
	errLocateItemSpan staticError = "locate item span"
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

type bodyItems struct {
	leading  hclwrite.Tokens
	items    []bodyItem
	trailing hclwrite.Tokens
}

type bodyItemPrefixes struct {
	leading  hclwrite.Tokens
	trailing hclwrite.Tokens
}

func emptyBodyItems() bodyItems {
	return bodyItems{
		leading:  nil,
		items:    nil,
		trailing: nil,
	}
}

func (collection bodyItems) isEmpty() bool {
	return len(collection.items) == indexFirst
}

// Format applies terraformat rules to a Terraform/HCL document.
func Format(src []byte, cfg config.Config) ([]byte, error) {
	startPos := hcl.Pos{
		Line:   startLine,
		Column: startColumn,
		Byte:   startByte,
	}

	file, diags := hclwrite.ParseConfig(src, "", startPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("%w: %s", errParseConfig, diags.Error())
	}

	ctx := bodyContext{root: true, blockType: emptyString}

	err := rewriteBody(file.Body(), ctx, cfg)
	if err != nil {
		return nil, err
	}

	out := file.Bytes()
	if cfg.EnsureEOFNewline {
		out = ensureTrailingNewline(out)
	}

	return out, nil
}

func rewriteBody(
	body *hclwrite.Body,
	ctx bodyContext,
	cfg config.Config,
) error {
	err := rewriteChildBlocks(body, cfg)
	if err != nil {
		return err
	}

	collection, err := collectBodyItems(body)
	if err != nil {
		return err
	}

	if collection.isEmpty() {
		return nil
	}

	if shouldApplyOrdering(cfg, ctx) {
		sortBodyItems(collection.items, ctx, cfg)
	}

	newTokens := renderBody(
		collection.leading,
		collection.items,
		collection.trailing,
		ctx,
		cfg,
	)

	body.Clear()
	body.AppendUnstructuredTokens(newTokens)

	return nil
}

func rewriteChildBlocks(body *hclwrite.Body, cfg config.Config) error {
	// Rewrite nested blocks first to avoid losing structure after reordering.
	for _, block := range body.Blocks() {
		childCtx := bodyContext{
			root:      false,
			blockType: block.Type(),
		}

		err := rewriteBody(block.Body(), childCtx, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func shouldApplyOrdering(cfg config.Config, ctx bodyContext) bool {
	if cfg.EnforceAttributeOrder {
		return true
	}

	return cfg.EnforceBlockOrder && ctx.root
}

func collectBodyItems(body *hclwrite.Body) (bodyItems, error) {
	attrItems := collectAttributeItems(body.Attributes())
	blockItems := collectBlockItems(body.Blocks())
	items := make([]bodyItem, indexFirst, len(attrItems)+len(blockItems))
	items = append(items, attrItems...)
	items = append(items, blockItems...)

	if len(items) == indexFirst {
		return emptyBodyItems(), nil
	}

	bodyTokens := body.BuildTokens(nil)

	err := assignItemSpans(items, bodyTokens)
	if err != nil {
		return emptyBodyItems(), err
	}

	sortItemsByStart(items)
	prefixes := applyItemPrefixes(items, bodyTokens)

	return bodyItems{
		leading:  prefixes.leading,
		items:    items,
		trailing: prefixes.trailing,
	}, nil
}

func collectAttributeItems(attrs map[string]*hclwrite.Attribute) []bodyItem {
	attrNames := make(map[*hclwrite.Attribute]string, len(attrs))
	for name, attr := range attrs {
		attrNames[attr] = name
	}

	items := make([]bodyItem, indexFirst, len(attrs))
	for _, attr := range attrs {
		item := newBodyItem(
			itemAttribute,
			attr,
			nil,
			attrNames[attr],
			emptyString,
			attr.BuildTokens(nil),
		)
		items = append(items, item)
	}

	return items
}

func collectBlockItems(blocks []*hclwrite.Block) []bodyItem {
	items := make([]bodyItem, indexFirst, len(blocks))
	for _, block := range blocks {
		item := newBodyItem(
			itemBlock,
			nil,
			block,
			block.Type(),
			strings.Join(block.Labels(), "."),
			block.BuildTokens(nil),
		)
		items = append(items, item)
	}

	return items
}

func assignItemSpans(items []bodyItem, bodyTokens hclwrite.Tokens) error {
	for itemIndex := range items {
		span, ok := findTokenSpan(bodyTokens, items[itemIndex].tokens)
		if !ok {
			return errLocateItemSpan
		}

		items[itemIndex].start = span.start
		items[itemIndex].end = span.end
	}

	return nil
}

func sortItemsByStart(items []bodyItem) {
	sort.Slice(items, func(leftIndex, rightIndex int) bool {
		return items[leftIndex].start < items[rightIndex].start
	})

	for itemIndex := range items {
		items[itemIndex].origIndex = itemIndex
	}
}

func applyItemPrefixes(
	items []bodyItem,
	bodyTokens hclwrite.Tokens,
) bodyItemPrefixes {
	prevEnd := indexNotFound

	for itemIndex := range items {
		start := items[itemIndex].start
		items[itemIndex].prefix = bodyTokens[prevEnd+indexOffset : start]
		prevEnd = items[itemIndex].end
	}

	trailing := bodyTokens[prevEnd+indexOffset:]

	leading := items[indexFirst].prefix
	items[indexFirst].prefix = nil

	return bodyItemPrefixes{
		leading:  leading,
		trailing: trailing,
	}
}

func renderBody(
	leading hclwrite.Tokens,
	items []bodyItem,
	trailing hclwrite.Tokens,
	ctx bodyContext,
	cfg config.Config,
) hclwrite.Tokens {
	out := normalizeLeadingTokens(leading)

	for itemIndex, item := range items {
		insertBlank := shouldInsertBlankLine(items, itemIndex, ctx, cfg)
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

func shouldInsertBlankLine(
	items []bodyItem,
	index int,
	ctx bodyContext,
	cfg config.Config,
) bool {
	if index == indexFirst {
		return false
	}

	if !cfg.EnforceTopLevelSpacing || !ctx.root {
		return false
	}

	prevIsBlock := items[index-indexOffset].kind == itemBlock
	currentIsBlock := items[index].kind == itemBlock

	return prevIsBlock && currentIsBlock
}

func ensureTrailingNewline(src []byte) []byte {
	if len(src) == indexFirst || bytes.HasSuffix(src, []byte("\n")) {
		return src
	}

	return append(src, '\n')
}

func newBodyItem(
	kind itemKind,
	attr *hclwrite.Attribute,
	block *hclwrite.Block,
	name string,
	labelKey string,
	tokens hclwrite.Tokens,
) bodyItem {
	return bodyItem{
		kind:      kind,
		attr:      attr,
		block:     block,
		name:      name,
		labelKey:  labelKey,
		tokens:    tokens,
		prefix:    nil,
		origIndex: indexFirst,
		start:     indexFirst,
		end:       indexFirst,
	}
}
