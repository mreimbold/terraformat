package format

import (
	"sort"

	"github.com/mreimbold/terraformat/config"
)

type sortKey struct {
	group int
	order int
	name  string
	label string
	index int
}

const (
	sortGroupDefault = iota
)

const (
	sortOrderDefault = iota
)

const (
	rootGroupAttributes = iota
	rootGroupBlocks
)

const (
	rootAttrOrderDefault = iota
)

type topLevelOrder int

const (
	topOrderTerraform topLevelOrder = iota
	topOrderProvider
	topOrderVariable
	topOrderLocals
	topOrderData
	topOrderResource
	topOrderModule
	topOrderOutput
	topOrderMoved
	topOrderImport
	topOrderCheck
	topOrderAssert
)

const topOrderUnknown topLevelOrder = 100

func sortBodyItems(items []bodyItem, ctx bodyContext, cfg config.Config) {
	for itemIndex := range items {
		items[itemIndex].origIndex = itemIndex
	}

	sort.SliceStable(items, func(leftIndex, rightIndex int) bool {
		left := itemSortKey(items[leftIndex], ctx, cfg)
		right := itemSortKey(items[rightIndex], ctx, cfg)

		return lessSortKey(left, right)
	})
}

func lessSortKey(left sortKey, right sortKey) bool {
	if left.group != right.group {
		return left.group < right.group
	}

	if left.order != right.order {
		return left.order < right.order
	}

	if left.name != right.name {
		return left.name < right.name
	}

	if left.label != right.label {
		return left.label < right.label
	}

	return left.index < right.index
}

func itemSortKey(item bodyItem, ctx bodyContext, cfg config.Config) sortKey {
	if ctx.root {
		return rootSortKey(item, cfg)
	}

	sorter := blockSorter(ctx.blockType)
	if sorter == nil {
		return defaultSortKey(item)
	}

	return sorter(item)
}

func blockSorter(blockType string) func(bodyItem) sortKey {
	sorters := map[string]func(bodyItem) sortKey{
		"resource":  resourceSortKey,
		"data":      resourceSortKey,
		"variable":  variableSortKey,
		"output":    outputSortKey,
		"module":    moduleSortKey,
		"provider":  providerSortKey,
		"terraform": terraformSortKey,
		"locals":    localsSortKey,
		"lifecycle": lifecycleSortKey,
	}

	return sorters[blockType]
}

func rootSortKey(item bodyItem, cfg config.Config) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		key.group = rootGroupAttributes

		if cfg.EnforceAttributeOrder {
			key.order = rootAttrOrderDefault
			key.name = item.name

			return key
		}

		key.order = item.origIndex

		return key
	}

	key.group = rootGroupBlocks

	if cfg.EnforceBlockOrder {
		order, ok := topLevelBlockOrder()[item.name]
		if ok {
			key.order = int(order)
		} else {
			key.order = int(topOrderUnknown)
		}

		key.name = item.name
		key.label = item.labelKey

		return key
	}

	key.order = item.origIndex

	return key
}

func newSortKey(index int) sortKey {
	return sortKey{
		group: sortGroupDefault,
		order: sortOrderDefault,
		name:  "",
		label: "",
		index: index,
	}
}

func topLevelBlockOrder() map[string]topLevelOrder {
	return map[string]topLevelOrder{
		"terraform": topOrderTerraform,
		"provider":  topOrderProvider,
		"variable":  topOrderVariable,
		"locals":    topOrderLocals,
		"data":      topOrderData,
		"resource":  topOrderResource,
		"module":    topOrderModule,
		"output":    topOrderOutput,
		"moved":     topOrderMoved,
		"import":    topOrderImport,
		"check":     topOrderCheck,
		"assert":    topOrderAssert,
	}
}
