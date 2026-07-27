// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import (
	"cmp"
	"errors"
	"fmt"
	"net"

	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type NetworkConfig struct {
	// Name for the network (defaults to "packer-{{uuid}}-network").
	// If specified, Packer will use existing network.
	NetworkName string `mapstructure:"network_name" required:"false"`
	// Name for the subnet (defaults to "packer-{{uuid}}-subnet").
	// If specified, Packer will use existing subnet.
	SubnetName string `mapstructure:"subnet_name" required:"false"`
	// Subnet CIDR (defaults to "192.168.0.0/16").
	SubnetCidr string `mapstructure:"subnet_cidr" required:"false"`
	// Use external address for connection to virtual machine from internet (defaults to "false").
	UseExternalAddress bool `mapstructure:"use_external_address" required:"false"`
	// External address name (defaults to "packer-{{uuid}}-external-address").
	// Can be specified only if external address usage is enabled.
	ExternalAddressName string `mapstructure:"external_address_name" required:"false"`
	// Enables virtual machine ip conversion from ipv4 to ipv6 with RFC 6052 (defaults to "false").
	// Meant to be used when packer is in ipv6 only network.
	Nat64Enable bool `mapstructure:"nat64_enable" required:"false"`
	// Prefix used in nat64 conversion (defaults to "64:ff9b::/96" (RFC 6052 Well-Known Prefix)).
	// CIDR notation only.
	Nat64IPV6Prefix string `mapstructure:"nat64_ipv6_prefix" required:"false"`
}

func (c *NetworkConfig) SetDefaults() {
	c.SubnetCidr = cmp.Or(c.SubnetCidr, DefaultSubnetCidr)
	c.Nat64IPV6Prefix = cmp.Or(c.Nat64IPV6Prefix, DefaultIPV6Prefix)
}

func (c *NetworkConfig) Validate() error {
	var err error
	if _, parseErr := cidraddress.ParseCIDR4AddressString(c.SubnetCidr); parseErr != nil {
		err = errors.Join(err, fmt.Errorf("parse subnet CIDR: %w", parseErr))
	}
	if c.SubnetName != "" && c.NetworkName == "" {
		err = errors.Join(err, consterr.Error("when subnet_name is provided, network_name must be provided"))
	}
	if !c.UseExternalAddress && c.SubnetName == "" {
		err = errors.Join(err, consterr.Error("when use_external_address is false, subnet_name must be provided"))
	}
	if !c.UseExternalAddress && c.ExternalAddressName != "" {
		err = errors.Join(err, consterr.Error("when use_external_address is false, external_address_name must not be provided"))
	}
	if _, _, parseErr := net.ParseCIDR(c.Nat64IPV6Prefix); parseErr != nil {
		err = errors.Join(err, fmt.Errorf("parse nat64_ipv6_prefix CIDR: %w", parseErr))
	}
	return err
}
