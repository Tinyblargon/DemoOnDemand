package session

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

type Config struct {
	InsecureFlag    bool
	Debug           bool
	Persist         bool
	User            string
	Password        string
	VSphereServer   string
	DebugPath       string
	DebugPathRun    string
	VimSessionPath  string
	RestSessionPath string
	KeepAlive       int
	APITimeout      time.Duration
}

type Client struct {
	// The VIM/govmomi client.
	VimClient *govmomi.Client
}

var defaultAPITimeout = time.Minute * 5

func New(VMware *dod.VMwareConfiguration) (*Client, error) {
	sessionConfig := &Config{
		User:            VMware.User,
		Password:        VMware.Password,
		InsecureFlag:    VMware.Insecure,
		VSphereServer:   VMware.URL,
		Debug:           true,
		DebugPathRun:    "",
		DebugPath:       "",
		Persist:         false,
		VimSessionPath:  "",
		RestSessionPath: "",
		KeepAlive:       100,
		APITimeout:      defaultAPITimeout,
	}
	c, err := sessionConfig.client()
	return c, err
}

func (c *Config) client() (*Client, error) {
	client := new(Client)

	u, err := c.vimURL()
	if err != nil {
		return nil, fmt.Errorf("Error generating SOAP endpoint url: %s", err)
	}

	// Set up the VIM/govmomi client connection, or load a previous session
	client.VimClient, err = c.SavedVimSessionOrNew(u)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Config) vimURL() (*url.URL, error) {
	u, err := url.Parse("https://" + c.VSphereServer + "/sdk")
	if err != nil {
		return nil, fmt.Errorf("Error parse url: %s", err)
	}

	u.User = url.UserPassword(c.User, c.Password)

	return u, nil
}

func (c *Config) SavedVimSessionOrNew(u *url.URL) (*govmomi.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	client, err := newClientWithKeepAlive(ctx, u, c.InsecureFlag, c.KeepAlive)
	if err != nil {
		return nil, fmt.Errorf("error setting up new vSphere SOAP client: %s", err)
	}
	return client, nil
}

func newClientWithKeepAlive(ctx context.Context, u *url.URL, insecure bool, keepAlive int) (*govmomi.Client, error) {
	soapClient := soap.NewClient(u, insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	k := session.KeepAlive(c.Client.RoundTripper, time.Duration(keepAlive)*time.Minute)
	c.Client.RoundTripper = k

	// Only login if the URL contains user information.
	if u.User != nil {
		err = c.Login(ctx, u.User)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
