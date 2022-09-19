package template

import (
	"fmt"
	"net"
)

type Network struct {
	Name         string `json:"name" yaml:"name"`
	RouterSubnet string `json:"router-cidr" yaml:"router-cidr"`
}

func (n *Network) ValidateRouterCIDR() error {
	_, _, err := net.ParseCIDR(n.RouterSubnet)
	if err != nil {
		return fmt.Errorf("%v is not an valid router CIDR", n.RouterSubnet)
	}
	return nil
}
