package template

import (
	"fmt"
	"net"
)

type Network struct {
	Name   string `json:"name" yaml:"name"`
	Subnet string `json:"subnet" yaml:"subnet"`
}

func (n *Network) ValidateSubnet() error {
	_, _, err := net.ParseCIDR(n.Subnet)
	if err != nil {
		return fmt.Errorf("%v is not an valid subnet.", n.Subnet)
	}
	return nil
}
