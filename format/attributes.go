package format

func resourceSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "count":
			key.group = 0
			key.order = 0
		case "for_each":
			key.group = 0
			key.order = 1
		case "provider":
			key.group = 0
			key.order = 2
		case "depends_on":
			key.group = 4
			key.order = 0
		default:
			key.group = 1
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		if item.name == "lifecycle" {
			key.group = 3
			return key
		}
		key.group = 2
		key.name = item.name
		key.label = item.labelKey
		return key
	}

	return key
}

func variableSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "type":
			key.group = 0
			key.order = 0
		case "description":
			key.group = 0
			key.order = 1
		case "default":
			key.group = 0
			key.order = 2
		case "sensitive":
			key.group = 0
			key.order = 3
		case "nullable":
			key.group = 0
			key.order = 4
		default:
			key.group = 0
			key.order = 10
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		if item.name == "validation" {
			key.order = 0
		} else {
			key.order = 1
			key.name = item.name
			key.label = item.labelKey
		}
		return key
	}

	return key
}

func outputSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "value":
			key.group = 0
			key.order = 0
		case "description":
			key.group = 0
			key.order = 1
		case "sensitive":
			key.group = 0
			key.order = 2
		case "depends_on":
			key.group = 2
			key.order = 0
		default:
			key.group = 0
			key.order = 10
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		key.name = item.name
		key.label = item.labelKey
		return key
	}

	return key
}

func moduleSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "source":
			key.group = 0
			key.order = 0
		case "version":
			key.group = 0
			key.order = 1
		case "providers":
			key.group = 0
			key.order = 2
		case "count":
			key.group = 0
			key.order = 3
		case "for_each":
			key.group = 0
			key.order = 4
		case "depends_on":
			key.group = 2
			key.order = 0
		default:
			key.group = 0
			key.order = 10
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		key.name = item.name
		key.label = item.labelKey
		return key
	}

	return key
}

func providerSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "alias":
			key.group = 0
			key.order = 0
		default:
			key.group = 0
			key.order = 1
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		key.name = item.name
		key.label = item.labelKey
		return key
	}

	return key
}

func terraformSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "required_version":
			key.group = 0
			key.order = 0
		default:
			key.group = 0
			key.order = 1
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		switch item.name {
		case "required_providers":
			key.order = 0
		case "backend":
			key.order = 1
		case "cloud":
			key.order = 2
		default:
			key.order = 10
			key.name = item.name
			key.label = item.labelKey
		}
		return key
	}

	return key
}

func localsSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		key.group = 0
		key.name = item.name
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		key.name = item.name
		key.label = item.labelKey
		return key
	}

	return key
}

func lifecycleSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		switch item.name {
		case "create_before_destroy":
			key.group = 0
			key.order = 0
		case "prevent_destroy":
			key.group = 0
			key.order = 1
		case "ignore_changes":
			key.group = 0
			key.order = 2
		case "replace_triggered_by":
			key.group = 0
			key.order = 3
		default:
			key.group = 0
			key.order = 10
			key.name = item.name
		}
		return key
	}

	if item.kind == itemBlock {
		key.group = 1
		key.name = item.name
		key.label = item.labelKey
		return key
	}

	return key
}

func defaultSortKey(item bodyItem) sortKey {
	key := sortKey{index: item.origIndex}
	if item.kind == itemAttribute {
		key.group = 0
		key.name = item.name
		return key
	}
	if item.kind == itemBlock {
		key.group = 1
		key.name = item.name
		key.label = item.labelKey
		return key
	}
	return key
}
