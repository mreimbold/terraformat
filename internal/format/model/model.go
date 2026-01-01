// Package model defines shared formatting data structures.
package model

import "github.com/hashicorp/hcl/v2/hclwrite"

// StartLine is the default line used when parsing HCL.
const StartLine = 1

// StartColumn is the default column used when parsing HCL.
const StartColumn = 1

// StartByte is the default byte offset used when parsing HCL.
const StartByte = 0

// EmptyString is a shared empty string sentinel.
const EmptyString = ""

// IndexFirst is the first index value used in token lists.
const IndexFirst = 0

// IndexOffset is the offset used for advancing indexes.
const IndexOffset = 1

// IndexNotFound is the sentinel returned when no match is found.
const IndexNotFound = -1

// ItemKind identifies whether an item is an attribute or a block.
type ItemKind int

const (
	// ItemAttribute marks an attribute item.
	ItemAttribute ItemKind = iota
	// ItemBlock marks a block item.
	ItemBlock
)

// Context describes the current body scope.
type Context struct {
	Root      bool
	BlockType string
}

// Item stores the tokens and metadata for a body element.
type Item struct {
	Kind      ItemKind
	Attr      *hclwrite.Attribute
	Block     *hclwrite.Block
	Name      string
	LabelKey  string
	Tokens    hclwrite.Tokens
	Prefix    hclwrite.Tokens
	OrigIndex int
	Start     int
	End       int
}

// Items groups leading/trailing tokens with body items.
type Items struct {
	Leading  hclwrite.Tokens
	Items    []Item
	Trailing hclwrite.Tokens
}

// ItemPrefixes stores prefix tokens for items and trailing tokens.
type ItemPrefixes struct {
	Leading  hclwrite.Tokens
	Trailing hclwrite.Tokens
}

// EmptyItems returns an empty Items collection.
func EmptyItems() Items {
	return Items{
		Leading:  nil,
		Items:    nil,
		Trailing: nil,
	}
}

// IsEmpty reports whether the Items collection has no entries.
func (collection Items) IsEmpty() bool {
	return len(collection.Items) == IndexFirst
}
