package ordering

import (
	"sort"

	"github.com/mreimbold/terraformat/internal/config"
	"github.com/mreimbold/terraformat/internal/format/model"
)

type Key struct {
	Group int
	Order int
	Name  string
	Label string
	Index int
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

func SortItems(items []model.Item, ctx model.Context, cfg config.Config) {
	for itemIndex := range items {
		items[itemIndex].OrigIndex = itemIndex
	}

	sort.SliceStable(items, func(leftIndex, rightIndex int) bool {
		left := ItemSortKey(items[leftIndex], ctx, cfg)
		right := ItemSortKey(items[rightIndex], ctx, cfg)

		return lessKey(left, right)
	})
}

func lessKey(left Key, right Key) bool {
	if left.Group != right.Group {
		return left.Group < right.Group
	}

	if left.Order != right.Order {
		return left.Order < right.Order
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}

	if left.Label != right.Label {
		return left.Label < right.Label
	}

	return left.Index < right.Index
}

func ItemSortKey(item model.Item, ctx model.Context, cfg config.Config) Key {
	if ctx.Root {
		return rootSortKey(item, cfg)
	}

	sorter := blockSorter(ctx.BlockType)
	if sorter == nil {
		return defaultSortKey(item)
	}

	return sorter(item)
}

func blockSorter(blockType string) func(model.Item) Key {
	sorters := map[string]func(model.Item) Key{
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

func rootSortKey(item model.Item, cfg config.Config) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		key.Group = rootGroupAttributes

		if cfg.EnforceAttributeOrder {
			key.Order = rootAttrOrderDefault
			key.Name = item.Name

			return key
		}

		key.Order = item.OrigIndex

		return key
	}

	key.Group = rootGroupBlocks

	if cfg.EnforceBlockOrder {
		order, ok := topLevelBlockOrder()[item.Name]
		if ok {
			key.Order = int(order)
		} else {
			key.Order = int(topOrderUnknown)
		}

		if shouldSortRootLabels(item.Name) {
			key.Label = item.LabelKey
		}

		return key
	}

	key.Order = item.OrigIndex

	return key
}

func newKey(index int) Key {
	return Key{
		Group: sortGroupDefault,
		Order: sortOrderDefault,
		Name:  "",
		Label: "",
		Index: index,
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

func shouldSortRootLabels(blockType string) bool {
	switch blockType {
	case "variable", "output":
		return true
	default:
		return false
	}
}
