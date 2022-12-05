package networkinterfaceconfig

import (
	"net"
)

func Base() *[]string {
	return &[]string{
		"# This file describes the network interfaces available on your system",
		"# and how to activate them. For more information, see interfaces(5).",
		"",
		"source /etc/network/interfaces.d/*",
		"",
		"# The loopback network interface",
		"auto lo",
		"iface lo inet loopback",
		"",
		"# The primary network interface",
	}
}

func New(interfaceName string, address *net.IPNet, dhcp bool) []string {
	if dhcp {
		return newDHCP(interfaceName)
	}
	return newStatic(interfaceName, address)
}

func newDHCP(interfaceName string) []string {
	content := make([]string, 3)
	content[0] = "allow-hotplug " + interfaceName
	content[1] = "iface " + interfaceName + " inet dhcp"
	return content
}

func newStatic(interfaceName string, cidr *net.IPNet) []string {
	content := make([]string, 4)
	content[0] = "allow-hotplug " + interfaceName
	content[1] = "iface " + interfaceName + " inet static"
	content[2] = "  address " + cidr.String()
	return content
}
