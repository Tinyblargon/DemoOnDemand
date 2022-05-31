package global

import (
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
)

const Concurency uint = 5

var ConfigFolder string
var DataCenter string
var TemplateFodler string
var RouterFodler string
var DemoFodler string
var IngressVM string
var Pool string
var PostgreSQLConfig *programconfig.PostgreSQLConfiguration
var LDAPConfig *programconfig.LDAPConfiguration

func SetAll(config *programconfig.Configuration) {
	ConfigFolder = config.ConfigFolder

	DataCenter = config.VMware.DataCenter
	baseFolder := strings.Trim(config.VMware.DemoFolder, "/")
	TemplateFodler = baseFolder + "/Templates"
	RouterFodler = baseFolder + "/Router"
	DemoFodler = baseFolder + "/Demos"
	IngressVM = "routervm"
	Pool = config.VMware.Pool

	PostgreSQLConfig = &programconfig.PostgreSQLConfiguration{
		Database: config.PostgreSQL.Database,
		Password: config.PostgreSQL.Password,
		Host:     config.PostgreSQL.Host,
		User:     config.PostgreSQL.User,
		Port:     config.PostgreSQL.Port,
	}
	LDAPConfig = &programconfig.LDAPConfiguration{
		BindUser:     config.LDAP.BindUser,
		BindPassword: config.LDAP.BindPassword,
	}
}
