package networkinterfaceconfig

import (
	"net"
)

func Base() []string {
	content := make([]string, 10)
	content[0] = "# This file describes the network interfaces available on your system"
	content[1] = "# and how to activate them. For more information, see interfaces(5)."
	content[2] = ""
	content[3] = "source /etc/network/interfaces.d/*"
	content[4] = ""
	content[5] = "# The loopback network interface"
	content[6] = "auto lo"
	content[7] = "iface lo inet loopback"
	content[8] = ""
	content[9] = "# The primary network interface"
	return content
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
