package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Tinyblargon/DemoOnDemand/dod"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/session"
)

func main() {
	config := dod.GetConfigProgramConfig()

	c, err := session.New(&config.VMware)
	LogFatal(err)

	global.SetAll(config.VMware.DataCenter, config.VMware.DemoFolder, config.ConfigFolder)
	err = dod.Intialize(c.VimClient, global.DataCenter)

	portForward1 := &demo.PortForward{
		SourcePort:      1,
		DestinationPort: 2,
		DestinationIP:   "10.10.10.10",
	}
	portForward2 := &demo.PortForward{
		SourcePort:      1,
		DestinationPort: 2,
		DestinationIP:   "10.10.10.10",
	}
	portForward3 := &demo.PortForward{
		SourcePort:      1,
		DestinationPort: 2,
		DestinationIP:   "10.10.10.10",
	}
	portForwars := []*demo.PortForward{
		portForward1,
		portForward2,
		portForward3,
	}
	demoConfig := &demo.DemoConfig{
		PortForwards: portForwars,
	}

	_ = demoConfig
	err = demo.Import(c.VimClient, global.DataCenter, "/test-import/test", "demo-02", demoConfig)
	fmt.Println(err)

	err = demo.New(c.VimClient, global.DataCenter, "demo-02", "myusername", 2)
	fmt.Println(err)

	err = demo.Start(c.VimClient, global.DataCenter, "demo-02", "myusername", 2)
	fmt.Println(err)

	err = demo.Stop(c.VimClient, global.DataCenter, "demo-02", "myusername", 2)
	fmt.Println(err)

	err = demo.Delete(c.VimClient, global.DataCenter, "demo-02", "myusername", 2)
	fmt.Println(err)

	err = demo.DestroyTemplate(c.VimClient, global.DataCenter, "demo-02")
	fmt.Println(err)

	// err = folder.Delete(c.VimClient, config.VMware.DataCenter, "/testfolder/Templates/demo-01")

	fmt.Println("test")
	fmt.Println(err)
	os.Exit(0)
}

func LogFatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
