package format

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/mreimbold/terraformat/internal/config"
	"github.com/mreimbold/terraformat/internal/format/model"
	"github.com/mreimbold/terraformat/internal/format/ordering"
	"github.com/mreimbold/terraformat/internal/format/spacing"
	"github.com/mreimbold/terraformat/internal/format/tokens"
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

// Format applies terraformat rules to a Terraform/HCL document.
func Format(src []byte, cfg config.Config) ([]byte, error) {
	startPos := hcl.Pos{
		Line:   model.StartLine,
		Column: model.StartColumn,
		Byte:   model.StartByte,
	}

	file, diags := hclwrite.ParseConfig(src, "", startPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("%w: %s", errParseConfig, diags.Error())
	}

	ctx := model.Context{Root: true, BlockType: model.EmptyString}

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
	ctx model.Context,
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

	if collection.IsEmpty() {
		return nil
	}

	if shouldApplyOrdering(cfg, ctx) {
		ordering.SortItems(collection.Items, ctx, cfg)
	}

	newTokens := renderBody(
		collection.Leading,
		collection.Items,
		collection.Trailing,
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
		childCtx := model.Context{
			Root:      false,
			BlockType: block.Type(),
		}

		err := rewriteBody(block.Body(), childCtx, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func shouldApplyOrdering(cfg config.Config, ctx model.Context) bool {
	if cfg.EnforceAttributeOrder {
		return true
	}

	return cfg.EnforceBlockOrder && ctx.Root
}

func collectBodyItems(body *hclwrite.Body) (model.Items, error) {
	attrItems := collectAttributeItems(body.Attributes())
	blockItems := collectBlockItems(body.Blocks())
	items := make([]model.Item, model.IndexFirst, len(attrItems)+len(blockItems))
	items = append(items, attrItems...)
	items = append(items, blockItems...)

	if len(items) == model.IndexFirst {
		return model.EmptyItems(), nil
	}

	bodyTokens := body.BuildTokens(nil)

	err := assignItemSpans(items, bodyTokens)
	if err != nil {
		return model.EmptyItems(), err
	}

	sortItemsByStart(items)
	prefixes := applyItemPrefixes(items, bodyTokens)

	return model.Items{
		Leading:  prefixes.Leading,
		Items:    items,
		Trailing: prefixes.Trailing,
	}, nil
}

func collectAttributeItems(attrs map[string]*hclwrite.Attribute) []model.Item {
	attrNames := make(map[*hclwrite.Attribute]string, len(attrs))
	for name, attr := range attrs {
		attrNames[attr] = name
	}

	items := make([]model.Item, model.IndexFirst, len(attrs))
	for _, attr := range attrs {
		item := newItem(
			model.ItemAttribute,
			attr,
			nil,
			attrNames[attr],
			model.EmptyString,
			attr.BuildTokens(nil),
		)
		items = append(items, item)
	}

	return items
}

func collectBlockItems(blocks []*hclwrite.Block) []model.Item {
	items := make([]model.Item, model.IndexFirst, len(blocks))
	for _, block := range blocks {
		item := newItem(
			model.ItemBlock,
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

func assignItemSpans(items []model.Item, bodyTokens hclwrite.Tokens) error {
	for itemIndex := range items {
		span, ok := tokens.FindSpan(bodyTokens, items[itemIndex].Tokens)
		if !ok {
			return errLocateItemSpan
		}

		items[itemIndex].Start = span.Start
		items[itemIndex].End = span.End
	}

	return nil
}

func sortItemsByStart(items []model.Item) {
	sort.Slice(items, func(leftIndex, rightIndex int) bool {
		return items[leftIndex].Start < items[rightIndex].Start
	})

	for itemIndex := range items {
		items[itemIndex].OrigIndex = itemIndex
	}
}

func applyItemPrefixes(
	items []model.Item,
	bodyTokens hclwrite.Tokens,
) model.ItemPrefixes {
	prevEnd := model.IndexNotFound

	for itemIndex := range items {
		start := items[itemIndex].Start
		items[itemIndex].Prefix = bodyTokens[prevEnd+model.IndexOffset : start]
		prevEnd = items[itemIndex].End
	}

	trailing := bodyTokens[prevEnd+model.IndexOffset:]

	leading := items[model.IndexFirst].Prefix
	items[model.IndexFirst].Prefix = nil

	return model.ItemPrefixes{
		Leading:  leading,
		Trailing: trailing,
	}
}

func renderBody(
	leading hclwrite.Tokens,
	items []model.Item,
	trailing hclwrite.Tokens,
	ctx model.Context,
	cfg config.Config,
) hclwrite.Tokens {
	out := spacing.NormalizeLeadingTokens(leading)

	for itemIndex, item := range items {
		insertBlank := spacing.ShouldInsertBlankLine(items, itemIndex, ctx, cfg)
		if insertBlank {
			out = append(out, spacing.NewlineToken())
		}

		prefix := spacing.NormalizePrefixTokens(item.Prefix)
		if insertBlank && !spacing.ContainsComment(item.Prefix) {
			prefix = nil
		}

		out = append(out, prefix...)
		out = append(out, item.Tokens...)
	}

	out = append(out, trailing...)

	return out
}

func ensureTrailingNewline(src []byte) []byte {
	if len(src) == model.IndexFirst || bytes.HasSuffix(src, []byte("\n")) {
		return src
	}

	return append(src, '\n')
}

func newItem(
	kind model.ItemKind,
	attr *hclwrite.Attribute,
	block *hclwrite.Block,
	name string,
	labelKey string,
	tokens hclwrite.Tokens,
) model.Item {
	return model.Item{
		Kind:      kind,
		Attr:      attr,
		Block:     block,
		Name:      name,
		LabelKey:  labelKey,
		Tokens:    tokens,
		Prefix:    nil,
		OrigIndex: model.IndexFirst,
		Start:     model.IndexFirst,
		End:       model.IndexFirst,
	}
}
