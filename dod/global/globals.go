package global

import (
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
)

var ConfigFolder string
var DataCenter string
var TemplateFodler string
var RouterFodler string
var DemoFodler string
var IngressVM string
var PostgreSQLConfig *programconfig.PostgreSQLConfiguration

func SetAll(config *programconfig.Configuration) {
	ConfigFolder = config.ConfigFolder

	DataCenter = config.VMware.DataCenter
	baseFolder := strings.Trim(config.VMware.DemoFolder, "/")
	TemplateFodler = baseFolder + "/Templates"
	RouterFodler = baseFolder + "/Router"
	DemoFodler = baseFolder + "/Demos"
	IngressVM = "routervm"

	PostgreSQLConfig = &programconfig.PostgreSQLConfiguration{
		Database: config.PostgreSQL.Database,
		Password: config.PostgreSQL.Password,
		Host:     config.PostgreSQL.Host,
		User:     config.PostgreSQL.User,
		Port:     config.PostgreSQL.Port,
	}
}
