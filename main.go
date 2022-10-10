package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Tinyblargon/DemoOnDemand/dod/authentication"
	"github.com/Tinyblargon/DemoOnDemand/dod/authentication/backends/ldap"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/concurrency"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/logger"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vlan"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/backends/memory"
	_ "github.com/lib/pq"
)

func main() {
	config, err := programconfig.GetConfigProgramConfig()
	OutFatal(err)
	OutFatal(logger.Initialize(*config.LogPath))
	logger.Fatal(concurrency.Initialize(config.Concurrency.TreadsPerTask))
	scheduler.Main = NewSchedulerBackend(config.Concurrency.ConcurrentTasks)
	authentication.Main = NewAuthBackend(config.LDAP)
	db, err := database.New(*config.PostgreSQL)
	logger.Fatal(err)
	global.SetAll(config, db)

	logger.Fatal(vsphere.Initialize(config.VMware))
	logger.Fatal(vlan.Initialize(config.Vlan.Id, config.Vlan.Prefix))

	// err = dod.Intialize(c.VimClient, global.VMwareConfig.DataCenter)
	// c.VimClient.Logout()
	logger.Fatal(frontend.Initialize(config.API.SuperUser.User, config.API.SuperUser.Password, config.API.Token.Secret, config.API.Token.IssuerClaim, config.API.Token.ExpirationTime))
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
	validatedSettigns, err := ldap.New(&ldap.Settings{
		URL:                LDAPsettings.URL,
		BindDN:             LDAPsettings.BindDN,
		BindCredential:     LDAPsettings.BindPassword,
		InsecureSkipVerify: LDAPsettings.InsecureSkipVerify,
		UsernameAttribute:  LDAPsettings.UsernameAttribute,
		UserGroup: &ldap.Settings_Group{
			UsersDN: LDAPsettings.UserGroup.UsersDN,
		},
		AdminGroup: &ldap.Settings_Group{
			UsersDN: LDAPsettings.AdminGroup.UsersDN,
		},
	})
	logger.Fatal(err)
	return validatedSettigns
}
