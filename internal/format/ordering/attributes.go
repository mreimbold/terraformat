package ordering

import "github.com/mreimbold/terraformat/internal/format/model"

const (
	attrCount     = "count"
	attrDependsOn = "depends_on"
	attrForEach   = "for_each"
	attrProvider  = "provider"
)

const (
	resourceGroupMeta = iota
	resourceGroupAttributes
	resourceGroupBlocks
	resourceGroupLifecycle
	resourceGroupDependsOn
)

const (
	resourceOrderCount = iota
	resourceOrderForEach
	resourceOrderProvider
)

const (
	variableGroupAttributes = iota
	variableGroupBlocks
)

const (
	variableOrderType = iota
	variableOrderDescription
	variableOrderDefault
	variableOrderSensitive
	variableOrderNullable
)

const variableOrderOther = 100

const (
	variableBlockOrderValidation = iota
	variableBlockOrderOther
)

const (
	outputGroupAttributes = iota
	outputGroupBlocks
	outputGroupDependsOn
)

const (
	outputOrderDescription = iota
	outputOrderValue
	outputOrderSensitive
)

const outputOrderOther = 100

const (
	moduleGroupAttributes = iota
	moduleGroupBlocks
	moduleGroupDependsOn
)

const (
	moduleOrderSource = iota
	moduleOrderVersion
	moduleOrderProviders
	moduleOrderCount
	moduleOrderForEach
)

const moduleOrderOther = 100

const (
	providerGroupAttributes = iota
	providerGroupBlocks
)

const (
	providerOrderAlias = iota
	providerOrderOther
)

const (
	terraformGroupAttributes = iota
	terraformGroupBlocks
)

const (
	terraformOrderRequiredVersion = iota
	terraformOrderAttributeOther
)

const (
	terraformBlockOrderRequiredProviders = iota
	terraformBlockOrderBackend
	terraformBlockOrderCloud
)

const terraformBlockOrderOther = 100

const (
	localsGroupAttributes = iota
	localsGroupBlocks
)

const (
	lifecycleGroupAttributes = iota
	lifecycleGroupBlocks
)

const (
	lifecycleOrderCreateBeforeDestroy = iota
	lifecycleOrderPreventDestroy
	lifecycleOrderIgnoreChanges
	lifecycleOrderReplaceTriggeredBy
)

const lifecycleOrderOther = 100

const (
	defaultGroupAttributes = iota
	defaultGroupBlocks
)

func resourceSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case attrCount:
			key.Group = resourceGroupMeta
			key.Order = resourceOrderCount
		case attrForEach:
			key.Group = resourceGroupMeta
			key.Order = resourceOrderForEach
		case attrProvider:
			key.Group = resourceGroupMeta
			key.Order = resourceOrderProvider
		case attrDependsOn:
			key.Group = resourceGroupDependsOn
			key.Order = sortOrderDefault
		default:
			key.Group = resourceGroupAttributes
			key.Order = item.OrigIndex
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		if item.Name == "lifecycle" {
			key.Group = resourceGroupLifecycle

			return key
		}

		key.Group = resourceGroupBlocks
		key.Order = item.OrigIndex

		return key
	}

	return key
}

func variableSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case "type":
			key.Group = variableGroupAttributes
			key.Order = variableOrderType
		case "description":
			key.Group = variableGroupAttributes
			key.Order = variableOrderDescription
		case "default":
			key.Group = variableGroupAttributes
			key.Order = variableOrderDefault
		case "sensitive":
			key.Group = variableGroupAttributes
			key.Order = variableOrderSensitive
		case "nullable":
			key.Group = variableGroupAttributes
			key.Order = variableOrderNullable
		default:
			key.Group = variableGroupAttributes
			key.Order = variableOrderOther
			key.Name = item.Name
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = variableGroupBlocks
		if item.Name == "validation" {
			key.Order = variableBlockOrderValidation
		} else {
			key.Order = variableBlockOrderOther
			key.Name = item.Name
			key.Label = item.LabelKey
		}

		return key
	}

	return key
}

func outputSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case "description":
			key.Group = outputGroupAttributes
			key.Order = outputOrderDescription
		case "value":
			key.Group = outputGroupAttributes
			key.Order = outputOrderValue
		case "sensitive":
			key.Group = outputGroupAttributes
			key.Order = outputOrderSensitive
		case attrDependsOn:
			key.Group = outputGroupDependsOn
			key.Order = sortOrderDefault
		default:
			key.Group = outputGroupAttributes
			key.Order = outputOrderOther
			key.Name = item.Name
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = outputGroupBlocks
		key.Name = item.Name
		key.Label = item.LabelKey

		return key
	}

	return key
}

func moduleSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case "source":
			key.Group = moduleGroupAttributes
			key.Order = moduleOrderSource
		case "version":
			key.Group = moduleGroupAttributes
			key.Order = moduleOrderVersion
		case "providers":
			key.Group = moduleGroupAttributes
			key.Order = moduleOrderProviders
		case attrCount:
			key.Group = moduleGroupAttributes
			key.Order = moduleOrderCount
		case attrForEach:
			key.Group = moduleGroupAttributes
			key.Order = moduleOrderForEach
		case attrDependsOn:
			key.Group = moduleGroupDependsOn
			key.Order = sortOrderDefault
		default:
			key.Group = moduleGroupAttributes
			key.Order = moduleOrderOther
			key.Name = item.Name
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = moduleGroupBlocks
		key.Name = item.Name
		key.Label = item.LabelKey

		return key
	}

	return key
}

func providerSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case "alias":
			key.Group = providerGroupAttributes
			key.Order = providerOrderAlias
		default:
			key.Group = providerGroupAttributes
			key.Order = providerOrderOther
			key.Name = item.Name
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = providerGroupBlocks
		key.Name = item.Name
		key.Label = item.LabelKey

		return key
	}

	return key
}

func terraformSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case "required_version":
			key.Group = terraformGroupAttributes
			key.Order = terraformOrderRequiredVersion
		default:
			key.Group = terraformGroupAttributes
			key.Order = terraformOrderAttributeOther
			key.Name = item.Name
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = terraformGroupBlocks

		switch item.Name {
		case "required_providers":
			key.Order = terraformBlockOrderRequiredProviders
		case "backend":
			key.Order = terraformBlockOrderBackend
		case "cloud":
			key.Order = terraformBlockOrderCloud
		default:
			key.Order = terraformBlockOrderOther
			key.Name = item.Name
			key.Label = item.LabelKey
		}

		return key
	}

	return key
}

func localsSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		key.Group = localsGroupAttributes
		key.Order = item.OrigIndex

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = localsGroupBlocks
		key.Order = item.OrigIndex

		return key
	}

	return key
}

func lifecycleSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		switch item.Name {
		case "create_before_destroy":
			key.Group = lifecycleGroupAttributes
			key.Order = lifecycleOrderCreateBeforeDestroy
		case "prevent_destroy":
			key.Group = lifecycleGroupAttributes
			key.Order = lifecycleOrderPreventDestroy
		case "ignore_changes":
			key.Group = lifecycleGroupAttributes
			key.Order = lifecycleOrderIgnoreChanges
		case "replace_triggered_by":
			key.Group = lifecycleGroupAttributes
			key.Order = lifecycleOrderReplaceTriggeredBy
		default:
			key.Group = lifecycleGroupAttributes
			key.Order = lifecycleOrderOther
			key.Name = item.Name
		}

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = lifecycleGroupBlocks
		key.Name = item.Name
		key.Label = item.LabelKey

		return key
	}

	return key
}

func defaultSortKey(item model.Item) Key {
	key := newKey(item.OrigIndex)

	if item.Kind == model.ItemAttribute {
		key.Group = defaultGroupAttributes
		key.Order = item.OrigIndex

		return key
	}

	if item.Kind == model.ItemBlock {
		key.Group = defaultGroupBlocks
		key.Order = item.OrigIndex

		return key
	}

	return key
}
