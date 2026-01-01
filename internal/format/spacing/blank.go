package spacing

import (
	"github.com/mreimbold/terraformat/internal/config"
	"github.com/mreimbold/terraformat/internal/format/model"
	"github.com/mreimbold/terraformat/internal/format/ordering"
)

type group struct {
	group     int
	blockType string
	kind      model.ItemKind
}

// ShouldInsertBlankLine reports whether a blank line should be inserted.
func ShouldInsertBlankLine(
	items []model.Item,
	index int,
	ctx model.Context,
	cfg config.Config,
) bool {
	if index == model.IndexFirst {
		return false
	}

	if !cfg.EnforceTopLevelSpacing {
		return false
	}

	prev := items[index-model.IndexOffset]
	current := items[index]

	if ctx.Root {
		return prev.Kind == model.ItemBlock && current.Kind == model.ItemBlock
	}

	prevGroup := itemGroup(prev, ctx, cfg)
	currentGroup := itemGroup(current, ctx, cfg)

	if prevGroup.group != currentGroup.group {
		return true
	}

	if prev.Kind != model.ItemBlock || current.Kind != model.ItemBlock {
		return false
	}

	return prevGroup.blockType != currentGroup.blockType
}

func itemGroup(item model.Item, ctx model.Context, cfg config.Config) group {
	key := ordering.ItemSortKey(item, ctx, cfg)
	blockType := model.EmptyString

	if item.Kind == model.ItemBlock && item.Block != nil {
		blockType = item.Block.Type()
	}

	return group{
		group:     key.Group,
		blockType: blockType,
		kind:      item.Kind,
	}
}
