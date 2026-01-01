package format

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
	outputOrderValue = iota
	outputOrderDescription
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

func resourceSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case attrCount:
			key.group = resourceGroupMeta
			key.order = resourceOrderCount
		case attrForEach:
			key.group = resourceGroupMeta
			key.order = resourceOrderForEach
		case attrProvider:
			key.group = resourceGroupMeta
			key.order = resourceOrderProvider
		case attrDependsOn:
			key.group = resourceGroupDependsOn
			key.order = sortOrderDefault
		default:
			key.group = resourceGroupAttributes
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		if item.name == "lifecycle" {
			key.group = resourceGroupLifecycle

			return key
		}

		key.group = resourceGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}

func variableSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case "type":
			key.group = variableGroupAttributes
			key.order = variableOrderType
		case "description":
			key.group = variableGroupAttributes
			key.order = variableOrderDescription
		case "default":
			key.group = variableGroupAttributes
			key.order = variableOrderDefault
		case "sensitive":
			key.group = variableGroupAttributes
			key.order = variableOrderSensitive
		case "nullable":
			key.group = variableGroupAttributes
			key.order = variableOrderNullable
		default:
			key.group = variableGroupAttributes
			key.order = variableOrderOther
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		key.group = variableGroupBlocks
		if item.name == "validation" {
			key.order = variableBlockOrderValidation
		} else {
			key.order = variableBlockOrderOther
			key.name = item.name
			key.label = item.labelKey
		}

		return key
	}

	return key
}

func outputSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case "value":
			key.group = outputGroupAttributes
			key.order = outputOrderValue
		case "description":
			key.group = outputGroupAttributes
			key.order = outputOrderDescription
		case "sensitive":
			key.group = outputGroupAttributes
			key.order = outputOrderSensitive
		case attrDependsOn:
			key.group = outputGroupDependsOn
			key.order = sortOrderDefault
		default:
			key.group = outputGroupAttributes
			key.order = outputOrderOther
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		key.group = outputGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}

func moduleSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case "source":
			key.group = moduleGroupAttributes
			key.order = moduleOrderSource
		case "version":
			key.group = moduleGroupAttributes
			key.order = moduleOrderVersion
		case "providers":
			key.group = moduleGroupAttributes
			key.order = moduleOrderProviders
		case attrCount:
			key.group = moduleGroupAttributes
			key.order = moduleOrderCount
		case attrForEach:
			key.group = moduleGroupAttributes
			key.order = moduleOrderForEach
		case attrDependsOn:
			key.group = moduleGroupDependsOn
			key.order = sortOrderDefault
		default:
			key.group = moduleGroupAttributes
			key.order = moduleOrderOther
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		key.group = moduleGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}

func providerSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case "alias":
			key.group = providerGroupAttributes
			key.order = providerOrderAlias
		default:
			key.group = providerGroupAttributes
			key.order = providerOrderOther
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		key.group = providerGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}

func terraformSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case "required_version":
			key.group = terraformGroupAttributes
			key.order = terraformOrderRequiredVersion
		default:
			key.group = terraformGroupAttributes
			key.order = terraformOrderAttributeOther
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		key.group = terraformGroupBlocks

		switch item.name {
		case "required_providers":
			key.order = terraformBlockOrderRequiredProviders
		case "backend":
			key.order = terraformBlockOrderBackend
		case "cloud":
			key.order = terraformBlockOrderCloud
		default:
			key.order = terraformBlockOrderOther
			key.name = item.name
			key.label = item.labelKey
		}

		return key
	}

	return key
}

func localsSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		key.group = localsGroupAttributes
		key.name = item.name

		return key
	}

	if item.kind == itemBlock {
		key.group = localsGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}

func lifecycleSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		switch item.name {
		case "create_before_destroy":
			key.group = lifecycleGroupAttributes
			key.order = lifecycleOrderCreateBeforeDestroy
		case "prevent_destroy":
			key.group = lifecycleGroupAttributes
			key.order = lifecycleOrderPreventDestroy
		case "ignore_changes":
			key.group = lifecycleGroupAttributes
			key.order = lifecycleOrderIgnoreChanges
		case "replace_triggered_by":
			key.group = lifecycleGroupAttributes
			key.order = lifecycleOrderReplaceTriggeredBy
		default:
			key.group = lifecycleGroupAttributes
			key.order = lifecycleOrderOther
			key.name = item.name
		}

		return key
	}

	if item.kind == itemBlock {
		key.group = lifecycleGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}

func defaultSortKey(item bodyItem) sortKey {
	key := newSortKey(item.origIndex)

	if item.kind == itemAttribute {
		key.group = defaultGroupAttributes
		key.name = item.name

		return key
	}

	if item.kind == itemBlock {
		key.group = defaultGroupBlocks
		key.name = item.name
		key.label = item.labelKey

		return key
	}

	return key
}
