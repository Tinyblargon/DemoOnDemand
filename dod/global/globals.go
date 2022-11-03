package global

import (
	"database/sql"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
)

var ConfigFolder string
var TemplateFolder string
var RouterFolder string
var DemoFolder string
var IngressVM string
var PostgreSQLConfig *programconfig.PostgreSQLConfiguration

// var LDAPConfig *programconfig.LDAPConfiguration
var TaskHistoryDepth uint
var DB *sql.DB

var RouterConfiguration *programconfig.SSHConfiguration

func SetAll(config *programconfig.Configuration, db *sql.DB) {
	ConfigFolder = config.ConfigFolder
	baseFolder := strings.Trim(config.VMware.DemoFolder, "/")
	TemplateFolder = baseFolder + "/Templates"
	RouterFolder = baseFolder + "/Router"
	DemoFolder = baseFolder + "/Demos"
	IngressVM = "routervm"

	TaskHistoryDepth = config.TaskHistoryDepth
	RouterConfiguration = config.Router
	DB = db
}
