package programconfig

import (
	"fmt"

	"github.com/spf13/viper"
)

// Configurations exported
type Configuration struct {
	ConfigFolder     string
	Concurrency      *Concurrency
	TaskHistoryDepth uint
	Vlan             *Vlan
	API              *APIConfiguration
	VMware           *VMwareConfiguration
	PostgreSQL       *PostgreSQLConfiguration
	LDAP             *LDAPConfiguration
	Router           *SSHConfiguration
	LogPath          *Logging
}

type Concurrency struct {
	ConcurrentTasks uint
	TreadsPerTask   uint
}

type Vlan struct {
	Id     *[]uint
	Prefix string
}

type APIConfiguration struct {
	PathPrefix   string
	Port         uint16
	CookieSecret string
	SuperUser    *SuperUser
}

// DatabaseConfigurations exported
type VMwareConfiguration struct {
	URL        string
	User       string
	Password   string
	Insecure   bool
	DataCenter string
	DemoFolder string
	Pool       string
	Hosts      []string
	Vswitch    string
}

type PostgreSQLConfiguration struct {
	Host     string
	User     string
	Password string
	Database string
	Port     uint16
}

type LDAPConfiguration_Group struct {
	UsersDN string
	// LDAPFilter string
}

type LDAPConfiguration struct {
	URL                string
	BindDN             string
	BindPassword       string
	InsecureSkipVerify bool
	UsernameAttribute  string
	UserGroup          LDAPConfiguration_Group
	AdminGroup         LDAPConfiguration_Group
}

type SSHConfiguration struct {
	User     string
	Password string
	Port     uint16
}

type Logging struct {
	Access string
	Error  string
	Info   string
}

type SuperUser struct {
	User     string
	Password string
}

func GetConfigProgramConfig(path ...string) (configuration *Configuration, err error) {

	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	if len(path) == 0 {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(path[0])
	}

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Set undefined variables
	// viper.SetDefault("database.dbname", "test_db")

	err = viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
	return
}
