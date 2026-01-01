package model

import "github.com/hashicorp/hcl/v2/hclwrite"

const (
	StartLine   = 1
	StartColumn = 1
	StartByte   = 0
)

const EmptyString = ""

const (
	IndexFirst    = 0
	IndexOffset   = 1
	IndexNotFound = -1
)

type ItemKind int

const (
	ItemAttribute ItemKind = iota
	ItemBlock
)

type Context struct {
	Root      bool
	BlockType string
}

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

type Items struct {
	Leading  hclwrite.Tokens
	Items    []Item
	Trailing hclwrite.Tokens
}

type ItemPrefixes struct {
	Leading  hclwrite.Tokens
	Trailing hclwrite.Tokens
}

func EmptyItems() Items {
	return Items{
		Leading:  nil,
		Items:    nil,
		Trailing: nil,
	}
}

func (collection Items) IsEmpty() bool {
	return len(collection.Items) == IndexFirst
}
