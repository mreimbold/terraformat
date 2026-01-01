/**
 * # Network Module
 *
 * This module is used to create a virtual network with a subnet and a network security group.
 *
 */

terraform {
  required_version = ">= 1.3.9"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 4.0.0"
    }
    azurecaf = {
      source  = "aztfmod/azurecaf"
      version = ">= 1.2.28"
    }
  }
}

resource "azurerm_virtual_network" "this" {
  count = var.virtual_network == null ? 1 : 0

  name                = azurecaf_name.virtual_network[0].result
  address_space       = var.vnet_address_spaces
  location            = data.azurerm_resource_group.this.location
  resource_group_name = data.azurerm_resource_group.this.name
  tags                = var.tags
}

resource "azurerm_subnet" "this" {
  for_each = { for subnet in var.subnets : subnet.name => subnet }

  name                              = azurecaf_name.subnet[each.key].result
  resource_group_name               = data.azurerm_resource_group.this.name
  virtual_network_name              = var.virtual_network != null ? data.azurerm_virtual_network.this[0].name : azurerm_virtual_network.this[0].name
  address_prefixes                  = each.value.address_prefixes
  service_endpoints                 = each.value.service_endpoints
  private_endpoint_network_policies = each.value.private_endpoint_network_policies

  dynamic "delegation" {
    for_each = each.value.delegations
    content {
      name = delegation.value.name
      service_delegation {
        name    = delegation.value.service_delegation.name
        actions = delegation.value.service_delegation.actions
      }
    }
  }

  lifecycle {
    # Workaround:
    # Checking this on the azurerm_virtual_network resource wouldn't work,
    # because if a virtual_network as input variable is provided that wouldn't get created in the first place
    precondition {
      condition     = var.virtual_network == null || length(var.vnet_address_spaces) == 0
      error_message = "If a virtual network is provided, the vnet_address_spaces do not take effect."
    }
  }
}

resource "azurerm_network_security_group" "this" {
  for_each = { for nsg in var.network_security_groups : nsg.name => nsg }

  name                = azurecaf_name.network_security_group[each.key].result
  location            = data.azurerm_resource_group.this.location
  resource_group_name = data.azurerm_resource_group.this.name
  tags                = var.tags

  dynamic "security_rule" {
    for_each = each.value.security_rules
    content {
      name                       = security_rule.value.name
      priority                   = security_rule.value.priority
      direction                  = security_rule.value.direction
      access                     = security_rule.value.access
      protocol                   = security_rule.value.protocol
      source_port_range          = security_rule.value.source_port_range
      destination_port_range     = security_rule.value.destination_port_range
      source_address_prefix      = security_rule.value.source_address_prefix
      destination_address_prefix = security_rule.value.destination_address_prefix
    }
  }
}

resource "azurerm_subnet_network_security_group_association" "subnet_nsg_assoc" {
  for_each = { for nsg in var.network_security_groups : nsg.name => nsg if lookup(nsg, "subnet_name", null) != null }

  subnet_id                 = azurerm_subnet.this[each.value.subnet_name].id
  network_security_group_id = azurerm_network_security_group.this[each.key].id
}
