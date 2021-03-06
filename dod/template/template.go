package template

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/file"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"gopkg.in/yaml.v2"
)

type PortForward struct {
	SourcePort      uint   `json:"sourceport" yaml:"sourceport"`
	DestinationPort uint   `json:"destinationport,omitempty" yaml:"destinationport,omitempty"`
	DestinationIP   string `json:"destinationip" yaml:"destinationip"`
	Protocol        string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

type Config struct {
	Name         string              `json:"name,omitempty" yaml:"name,omitempty"`
	Description  string              `json:"description,omitempty" yaml:"description"`
	Path         string              `json:"path,omitempty" yaml:"path,omitempty"`
	PortForwards []*PortForward      `json:"portforwards" yaml:"portforwards"`
	Networks     []*template.Network `json:"networks" yaml:"networks"`
}

func Get(templateName string) (templateConfig *Config, err error) {
	contents, err := file.Read(global.ConfigFolder + "/" + templateName)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(contents, &templateConfig)
	return
}

func Destroy(client *govmomi.Client, dc *object.Datacenter, TempalateName string, status *taskstatus.Status) (err error) {
	inUse, err := database.CheckTemplateInUse(global.DB, TempalateName)
	if err != nil {
		return
	}
	if inUse {
		return fmt.Errorf("unable to remove template, template is in use")
	}
	err = folder.Delete(client, dc, global.TemplateFodler+"/"+TempalateName, status)
	if err != nil {
		return
	}
	return file.Delete(global.ConfigFolder + "/" + TempalateName)
}

func List() (files []string, err error) {
	return file.ReadDir(global.ConfigFolder)
}

// Imports a new demo from the speciefid folder
func (c *Config) Import(client *govmomi.Client, dc *object.Datacenter, pool string, status *taskstatus.Status) (err error) {
	filePath := global.ConfigFolder + "/" + c.Name
	err = folder.Clone(client, dc, nil, c.Path, global.TemplateFodler+"/"+c.Name, pool, true, status)
	if err != nil {
		return
	}
	return c.WriteToFile(filePath)
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

func (c *Config) WriteToFile(filePath string) error {
	c.Name = ""
	c.Path = ""
	data, _ := yaml.Marshal(c)
	return file.Write(filePath, data)
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
