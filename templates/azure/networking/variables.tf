variable "vpc_block" {
  type    = string
  default = "192.168.0.0/16"
}

variable "public_subnet_01_block" {
  type    = string
  default = "192.168.0.0/18"
}

variable "public_subnet_02_block" {
  type    = string
  default = "192.168.64.0/18"
}

variable "private_subnet_01_block" {
  type    = string
  default = "192.168.128.0/18"
}

variable "private_subnet_02_block" {
  type    = string
  default = "192.168.192.0/18"
}
