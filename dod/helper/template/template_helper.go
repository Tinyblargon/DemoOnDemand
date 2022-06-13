package template

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/file"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/vmware/govmomi"
	"gopkg.in/yaml.v2"
)

type PortForward struct {
	SourcePort      uint   `json:"sourceport"`
	DestinationPort uint   `json:"destinationport,omitempty"`
	DestinationIP   string `json:"destinationip"`
	Protocol        string `json:"protocol,omitempty"`
}

type Network struct {
	Name   string `json:"name"`
	Subnet string `json:"subnet"`
}

type Config struct {
	Name         string         `json:"name,omitempty"`
	Description  string         `json:"description,omitempty"`
	Path         string         `json:"path,omitempty"`
	PortForwards []*PortForward `json:"portforwards"`
	Networks     []*Network     `json:"networks"`
}

func Get(templateName string) (templateConfig *Config, err error) {
	contents, err := file.Read(global.ConfigFolder + "/" + templateName)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(contents, &templateConfig)
	return
}

// Imports a new demo from the speciefid folder
func (c *Config) Import(client *govmomi.Client, dataCenter, pool string, status *taskstatus.Status) (err error) {
	filePath := global.ConfigFolder + "/" + c.Name
	err = folder.Clone(client, dataCenter, c.Path, global.TemplateFodler+"/"+c.Name, pool, true, status)
	if err != nil {
		return
	}
	data, _ := yaml.Marshal(c)
	return file.Write(filePath, data)
}

func (c *Config) Defaults() {
	for _, e := range c.PortForwards {
		if e.DestinationPort == 0 {
			e.DestinationPort = e.SourcePort
		}
		if e.Protocol == "" {
			e.Protocol = "TCP"
		}
	}
}

func (c *Config) Validate(nameAndPathEmpty bool) (err error) {
	if !nameAndPathEmpty {
		if c.Name == "" {
			return fmt.Errorf("name may not be empty")
		}
		if c.Path == "" {
			return fmt.Errorf("path may not be empty")
		}
	}
	err = c.ValidatePortforwards()
	if err != nil {
		return
	}
	for _, e := range c.PortForwards {
		err = e.ValidateIP()
		if err != nil {
			return
		}
	}
	for _, e := range c.Networks {
		err = e.ValidateSubnet()
		if err != nil {
			return
		}
	}
	return err
}

func (c *Config) ValidatePortforwards() (err error) {
	list := make([]string, 0)
	for _, e := range c.PortForwards {
		err = ValidateSourcePort(e.SourcePort)
		if err != nil {
			return
		}
		err = ValidateDestinationPort(e.DestinationPort)
		if err != nil {
			return
		}
		err = ValidateProtocol(e.Protocol)
		if err != nil {
			return
		}
		item := e.Protocol + strconv.Itoa(int(e.SourcePort))
		if !util.IsStringUnique(&list, item) {
			return fmt.Errorf("duplicate sourceport.")
		}
		list = append(list, item)
	}
	return
}

func ValidateSourcePort(port uint) error {
	if port == 0 && port > 65353 {
		return fmt.Errorf("%d is not an valid sourceport.", port)
	}
	return nil
}

func ValidateDestinationPort(port uint) error {
	if port > 65353 {
		return fmt.Errorf("%d is not an valid destinationport.", port)
	}
	return nil
}

func ValidateProtocol(protocol string) error {
	if !strings.EqualFold(protocol, "TCP") && !strings.EqualFold(protocol, "UDP") {
		return fmt.Errorf("%s is not an valid protocol.", protocol)
	}
	return nil
}

func (p *PortForward) ValidateIP() error {
	trial := net.ParseIP(p.DestinationIP)
	if trial.To4() == nil && trial.To16() == nil {
		return fmt.Errorf("%v is not an valid IP address.", trial)
	}
	return nil
}

func (n *Network) ValidateSubnet() error {
	_, _, err := net.ParseCIDR(n.Subnet)
	if err != nil {
		return fmt.Errorf("%v is not an valid subnet.", n.Subnet)
	}
	return nil
}
