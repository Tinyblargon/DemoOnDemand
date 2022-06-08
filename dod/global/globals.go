package global

import (
	"database/sql"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
)

const Concurency uint = 5

var CookieSecret []byte = []byte("keymaker")

var ConfigFolder string
var TemplateFodler string
var RouterFodler string
var DemoFodler string
var IngressVM string
var PostgreSQLConfig *programconfig.PostgreSQLConfiguration

// var LDAPConfig *programconfig.LDAPConfiguration
var TaskHistoryDepth uint
var DB *sql.DB
var VMwareConfig *programconfig.VMwareConfiguration

func SetAll(config *programconfig.Configuration, db *sql.DB) {
	ConfigFolder = config.ConfigFolder
	baseFolder := strings.Trim(config.VMware.DemoFolder, "/")
	TemplateFodler = baseFolder + "/Templates"
	RouterFodler = baseFolder + "/Router"
	DemoFodler = baseFolder + "/Demos"
	IngressVM = "routervm"

	VMwareConfig = &programconfig.VMwareConfiguration{
		URL:        config.VMware.URL,
		User:       config.VMware.User,
		Password:   config.VMware.Password,
		Insecure:   config.VMware.Insecure,
		DataCenter: config.VMware.DataCenter,
		Pool:       config.VMware.Pool,
	}

	// PostgreSQLConfig = &programconfig.PostgreSQLConfiguration{
	// 	Database: config.PostgreSQL.Database,
	// 	Password: config.PostgreSQL.Password,
	// 	Host:     config.PostgreSQL.Host,
	// 	User:     config.PostgreSQL.User,
	// 	Port:     config.PostgreSQL.Port,
	// }
	// LDAPConfig = &programconfig.LDAPConfiguration{
	// 	BindUser:     config.LDAP.BindUser,
	// 	BindPassword: config.LDAP.BindPassword,
	// }
	TaskHistoryDepth = 50
	DB = db
}
