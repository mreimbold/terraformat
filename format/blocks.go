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

var topLevelBlockOrder = map[string]int{
	"terraform": 0,
	"provider":  1,
	"variable":  2,
	"locals":    3,
	"data":      4,
	"resource":  5,
	"module":    6,
	"output":    7,
	"moved":     8,
	"import":    9,
	"check":     10,
	"assert":    11,
}

func sortBodyItems(items []bodyItem, ctx bodyContext, cfg config.Config) {
	for i := range items {
		items[i].origIndex = i
	}

	sort.SliceStable(items, func(i, j int) bool {
		left := itemSortKey(items[i], ctx, cfg)
		right := itemSortKey(items[j], ctx, cfg)
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
	})
}

func itemSortKey(item bodyItem, ctx bodyContext, cfg config.Config) sortKey {
	if ctx.root {
		return rootSortKey(item, cfg)
	}

	switch ctx.blockType {
	case "resource", "data":
		return resourceSortKey(item)
	case "variable":
		return variableSortKey(item)
	case "output":
		return outputSortKey(item)
	case "module":
		return moduleSortKey(item)
	case "provider":
		return providerSortKey(item)
	case "terraform":
		return terraformSortKey(item)
	case "locals":
		return localsSortKey(item)
	case "lifecycle":
		return lifecycleSortKey(item)
	default:
		return defaultSortKey(item)
	}
}

func rootSortKey(item bodyItem, cfg config.Config) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		key.group = 0
		if cfg.EnforceAttributeOrder {
			key.order = 0
			key.name = item.name
			return key
		}
		key.order = item.origIndex
		return key
	}

	key.group = 1
	if cfg.EnforceBlockOrder {
		if order, ok := topLevelBlockOrder[item.name]; ok {
			key.order = order
		} else {
			key.order = 100
		}
		key.name = item.name
		key.label = item.labelKey
		return key
	}
	key.order = item.origIndex
	return key
}
