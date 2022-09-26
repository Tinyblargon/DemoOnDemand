package firewallconfig

import (
	"strconv"
	"strings"
)

const FirewallFile string = "/etc/network/if-pre-up.d/iptables"

func Base(mainInterface string) *[]string {
	return &[]string{
		"#!/bin/sh",
		"echo 1 > /proc/sys/net/ipv4/ip_forward",
		"iptables --flush",
		"iptables --table nat --flush",
		"iptables --delete-chain",
		"iptables --table nat --delete-chain",
		"iptables --policy INPUT DROP",
		"iptables --policy OUTPUT ACCEPT",
		"iptables --policy FORWARD ACCEPT",
		"iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT",
		"iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT",
		"iptables -A INPUT -i lo -j ACCEPT",
		"iptables -A OUTPUT -o lo -j ACCEPT",
		"iptables -A INPUT -p icmp --icmp-type echo-request -j ACCEPT",
		"iptables --table nat --append POSTROUTING --out-interface " + mainInterface + " -j MASQUERADE",
	}
}

func New(protocol string, port uint16) string {
	return "iptables -A INPUT -p " + strings.ToLower(protocol) + " --dport " + strconv.Itoa(int(port)) + " -j ACCEPT"
}

func NewPrerouting(sourcePort, destinationPort uint16, ip, protocol, mainInterface string) string {
	return "iptables -A PREROUTING -t nat -i " + mainInterface + " -p " + strings.ToLower(protocol) + " --dport " + strconv.Itoa(int(sourcePort)) + " -j DNAT --to " + ip + ":" + strconv.Itoa(int(destinationPort))
}
