package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Tinyblargon/DemoOnDemand/dod/authentication"
	"github.com/Tinyblargon/DemoOnDemand/dod/authentication/backends/ldap"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/logger"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/host"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/backends/memory"
	_ "github.com/lib/pq"
)

func main() {
	config, err := programconfig.GetConfigProgramConfig()
	OutFatal(err)
	OutFatal(logger.Initialize(*config.LogPath))
	scheduler.Main = NewSchedulerBackend(config.ConcurrentTasks)
	authentication.Main = NewAuthBackend(config.LDAP)
	db, err := database.New(*config.PostgreSQL)
	logger.Fatal(err)
	global.SetAll(config, db)

	c, err := session.New(*config.VMware)
	logger.Fatal(err)
	logger.Fatal(datacenter.Initialize(c.VimClient, global.VMwareConfig.DataCenter))
	logger.Fatal(host.Initialize(c.VimClient, datacenter.GetObject(), global.VMwareConfig.Hosts))

	vlan.Initialize(config.Vlan.Id, config.Vlan.Prefix)

	// err = dod.Intialize(c.VimClient, global.VMwareConfig.DataCenter)
	// c.VimClient.Logout()
	logger.Fatal(frontend.Initialize(config.API.SuperUser.User, config.API.SuperUser.Password, config.API.CookieSecret))
	frontend.HandleRequests(config.API.PathPrefix, config.API.Port)

	db.Close()
	fmt.Println(err)
	os.Exit(0)
}

func OutFatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func NewSchedulerBackend(concurrency uint) scheduler.Backend {
	return memory.New(concurrency)
}

func NewAuthBackend(LDAPsettings *programconfig.LDAPConfiguration) authentication.Backend {
	userGroup := ldap.Settings_Group{
		UsersDN: LDAPsettings.UserGroup.UsersDN,
	}
	adminGroup := ldap.Settings_Group{
		UsersDN: LDAPsettings.AdminGroup.UsersDN,
	}
	settings := ldap.Settings{
		URL:                LDAPsettings.URL,
		BindDN:             LDAPsettings.BindDN,
		BindCredential:     LDAPsettings.BindPassword,
		InsecureSkipVerify: LDAPsettings.InsecureSkipVerify,
		UsernameAttribute:  LDAPsettings.UsernameAttribute,
		UserGroup:          &userGroup,
		AdminGroup:         &adminGroup,
	}
	validatedSettigns, err := ldap.New(&settings)
	logger.Fatal(err)
	return validatedSettigns
}
