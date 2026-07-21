// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"fmt"
	"net"

	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
)

// ConvertToIPv6 converts an ipaddress.IPAddress containing an IPv4 address
// to an IPv6 address according to RFC 6052.
// https://www.rfc-editor.org/info/rfc6052/
func ConvertToIPv6(ipv4 net.IP, prefixStr string) (*ipaddress.IPAddress, error) {
	if ipv4 == nil {
		return nil, fmt.Errorf("ipaddress is nil")
	}

	ipv4 = ipv4.To4()
	if ipv4 == nil {
		return nil, fmt.Errorf("invalid IPv4 address")
	}

	_, prefix, err := net.ParseCIDR(prefixStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NAT64 prefix: %w", err)
	}
	ipv6, err := synthesizeRFC6052(*prefix, ipv4)
	if err != nil {
		return nil, fmt.Errorf("failed to convert with RFC6052: %w", err)
	}

	result, err := ipaddress.NewIPAddress(ipv6)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// synthesizeRFC6052 converts an IPv4 address into an IPv6 address using RFC 6052 rules
// for any valid prefix length (32, 40, 48, 56, 64, or 96).
func synthesizeRFC6052(prefix net.IPNet, ipv4 net.IP) (net.IP, error) {
	ipv6 := make(net.IP, net.IPv6len)

	ones, _ := prefix.Mask.Size()
	pos := ones / 8

	// Embed the IPv4 address based on RFC 6052 mapping rules
	copy(ipv6, prefix.IP)
	copy(ipv6[pos:], ipv4)
	// Bits 64-71 (v6[8]) are reserved and MUST be 0
	switch ones {
	case 32, 40, 48, 56, 64:
		copy(ipv6[9:], ipv6[8:])
	case 96:
		// Ensure that bits 64 to 71 are set to zero
		ipv6[8] = 0
	default:
		return nil, fmt.Errorf("unsupported RFC 6052 prefix length: %d", ones)
	}

	return ipv6, nil
}
