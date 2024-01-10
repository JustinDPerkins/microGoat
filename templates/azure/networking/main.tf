# This template Create a Base Virtual Network for the Dunder Mifflen App. 2 public, 2 private subnets. 2 NAT gateways.
# Create resource group
resource "azurerm_resource_group" "dunder-rg" {
  name     = "dunder-rg"
  location = "East US"
}

# Create VNET for Dunder Mifflen
resource "azurerm_virtual_network" "dunder-vpc" {
  name                = "dunder-vnet"
  address_space       = [var.vpc_block]
  location            = azurerm_resource_group.dunder-rg.location
  resource_group_name = azurerm_resource_group.dunder-rg.name
}

# Create public subnet 1
resource "azurerm_subnet" "public_subnet_01" {
  name                 = "dunder-public-subnet-01"
  resource_group_name  = azurerm_resource_group.dunder-rg.name
  virtual_network_name = azurerm_virtual_network.dunder-vpc.name
  address_prefixes     = [var.public_subnet_01_block]
}

# Create public subnet 2
resource "azurerm_subnet" "public_subnet_02" {
  name                 = "dunder-public-subnet-02"
  resource_group_name  = azurerm_resource_group.dunder-rg.name
  virtual_network_name = azurerm_virtual_network.dunder-vpc.name
  address_prefixes     = [var.public_subnet_02_block]
}

# Create private subnet 1
resource "azurerm_subnet" "private_subnet_01" {
  name                 = "dunder-private-subnet-01"
  resource_group_name  = azurerm_resource_group.dunder-rg.name
  virtual_network_name = azurerm_virtual_network.dunder-vpc.name
  address_prefixes     = [var.private_subnet_01_block]
}

# Create private subnet 2
resource "azurerm_subnet" "private_subnet_02" {
  name                 = "dunder-private-subnet-02"
  resource_group_name  = azurerm_resource_group.dunder-rg.name
  virtual_network_name = azurerm_virtual_network.dunder-vpc.name
  address_prefixes     = [var.private_subnet_02_block]
}

# Create public IPs for NAT gateways 1
resource "azurerm_public_ip" "nat_gateway_eip_01" {
  name                = "dunder-nat-gateway-eip-01"
  location            = azurerm_resource_group.dunder-rg.location
  resource_group_name = azurerm_resource_group.dunder-rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

# Create NAT gateway 1
resource "azurerm_nat_gateway" "nat_gateway_01" {
  name                = "dunder-nat-gateway-01"
  resource_group_name = azurerm_resource_group.dunder-rg.name
  location            = azurerm_resource_group.dunder-rg.location
}

# nat gateway 1 IP associate
resource "azurerm_nat_gateway_public_ip_association" "nat_gateway_01_association" {
  nat_gateway_id       = azurerm_nat_gateway.nat_gateway_01.id
  public_ip_address_id = azurerm_public_ip.nat_gateway_eip_01.id
}
# Assign Nat 1 to Pub Sub 1
resource "azurerm_subnet_nat_gateway_association" "nat_sub_1" {
  subnet_id      = azurerm_subnet.public_subnet_01.id
  nat_gateway_id = azurerm_nat_gateway.nat_gateway_01.id
}

# Nat Gateway 2 Pub IP
resource "azurerm_public_ip" "nat_gateway_eip_02" {
  name                = "dunder-nat-gateway-eip-02"
  location            = azurerm_resource_group.dunder-rg.location
  resource_group_name = azurerm_resource_group.dunder-rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

# create nat gtwy 2
resource "azurerm_nat_gateway" "nat_gateway_02" {
  name                = "dunder-nat-gateway-02"
  resource_group_name = azurerm_resource_group.dunder-rg.name
  location            = azurerm_resource_group.dunder-rg.location
}

# nat gateway 2 IP associate
resource "azurerm_nat_gateway_public_ip_association" "nat_gateway_02_association" {
  nat_gateway_id       = azurerm_nat_gateway.nat_gateway_02.id
  public_ip_address_id = azurerm_public_ip.nat_gateway_eip_02.id
}

# Assign Nat 2 to Pub Sub 2
resource "azurerm_subnet_nat_gateway_association" "nat_sub_2" {
  subnet_id      = azurerm_subnet.public_subnet_02.id
  nat_gateway_id = azurerm_nat_gateway.nat_gateway_02.id
}

# Create route table for private routes
resource "azurerm_route_table" "private_route_table" {
  name                = "private-route-table"
  resource_group_name = azurerm_resource_group.dunder-rg.name
  location            = azurerm_resource_group.dunder-rg.location
}

# Create default route pointing to NAT gateway for private subnet 1
resource "azurerm_route" "private_subnet_route_01" {
  name                = "private-subnet-route-01"
  resource_group_name = azurerm_resource_group.dunder-rg.name
  route_table_name    = azurerm_route_table.private_route_table.name
  address_prefix      = var.private_subnet_01_block
  next_hop_type       = "VirtualAppliance"
  next_hop_in_ip_address = azurerm_public_ip.nat_gateway_eip_01.ip_address
}

# Create default route pointing to NAT gateway for private subnet 2
resource "azurerm_route" "private_subnet_route_02" {
  name                = "private-subnet-route-02"
  resource_group_name = azurerm_resource_group.dunder-rg.name
  route_table_name    = azurerm_route_table.private_route_table.name
  address_prefix      = var.private_subnet_02_block
  next_hop_type       = "VirtualAppliance"
  next_hop_in_ip_address = azurerm_public_ip.nat_gateway_eip_02.ip_address
}

# Create private routes association 1
resource "azurerm_subnet_route_table_association" "private_subnet_association_01" {
  subnet_id      = azurerm_subnet.private_subnet_01.id
  route_table_id = azurerm_route_table.private_route_table.id
}

# Create private routes association 2
resource "azurerm_subnet_route_table_association" "private_subnet_association_02" {
  subnet_id      = azurerm_subnet.private_subnet_02.id
  route_table_id = azurerm_route_table.private_route_table.id
}

output "public_subnet_01" {
  value = azurerm_subnet.public_subnet_01.id
}

output "public_subnet_02" {
  value = azurerm_subnet.public_subnet_02.id
}

output "private_subnet_01" {
  value = azurerm_subnet.private_subnet_01.id
}

output "private_subnet_02" {
  value = azurerm_subnet.private_subnet_02.id
}

output "dunder_vpc_id" {
  value = azurerm_virtual_network.dunder-vpc.id
}
