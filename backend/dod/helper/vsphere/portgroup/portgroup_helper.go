package portgroup

import (
	"fmt"
	"strconv"

	"context"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/concurrency"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/name"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const itemDoesNotExist string = "ServerFaultCode: The object or item referred to could not be found."

type Networks struct {
	Vlan uint
	Host *object.HostSystem
}

func Create(c *govmomi.Client, hosts []*object.HostSystem, vlans *[]uint, prefix, vSwitch string, concurrencyNumber uint, status *taskstatus.Status) error {
	numberOfTasks := uint(len(*vlans) * len(hosts))
	in, conObject := channelInitialize(numberOfTasks, concurrencyNumber)
	// spawn "conObject.Threads" amount of threads
	for x := 0; x < int(conObject.Threads); x++ {
		go func() {
			for x := range in {
				conObject.Cycle(createSingle(c, x.Host, prefix, vSwitch, x.Vlan, status))
				if conObject.Err != nil {
					break
				}
			}
		}()
	}
	return channelLooper(in, conObject, hosts, vlans)
}

// Creates a portgroup on a singular host
func createSingle(c *govmomi.Client, host *object.HostSystem, prefix, vSwitch string, vlan uint, status *taskstatus.Status) (err error) {
	ns, err := hostNetworkSystemFromHostSystem(host)
	if err != nil {
		return networkSystemError(err)
	}
	spec := expandHostPortGroupSpec(prefix, vSwitch, vlan)
	status.AddToInfo(fmt.Sprintf("Create portgroup %s on host %s", spec.Name, host.Name()))
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	err = ns.AddPortGroup(ctx, *spec)
	if err != nil {
		return fmt.Errorf("error adding port group: %s", err)
	}
	return
}

// expandHostPortGroupSpec reads certain ResourceData keys and returns a
// HostPortGroupSpec.
func expandHostPortGroupSpec(prefix, vSwitchName string, vlan uint) *types.HostPortGroupSpec {
	return &types.HostPortGroupSpec{
		Name:        name.Network(prefix, vlan),
		VlanId:      int32(vlan),
		VswitchName: vSwitchName,
		// Policy:      *expandHostNetworkPolicy(d),
	}
}

// ##############################################################################################

func Delete(c *govmomi.Client, hosts []*object.HostSystem, prefix string, vlans *[]uint, concurrencyNumber uint, status *taskstatus.Status) error {
	numberOfTasks := uint(len(*vlans) * len(hosts))
	in, conObject := channelInitialize(numberOfTasks, concurrencyNumber)
	// spawn "conObject.Threads" amount of threads
	for x := 0; x < int(conObject.Threads); x++ {
		go func() {
			for x := range in {
				conObject.Cycle(deleteSingle(c, x.Host, prefix, x.Vlan, status))
				if conObject.Err != nil {
					break
				}
			}
		}()
	}
	return channelLooper(in, conObject, hosts, vlans)
}

// Deletes a portGroup on a singular host
func deleteSingle(c *govmomi.Client, host *object.HostSystem, prefix string, vlan uint, status *taskstatus.Status) error {
	ns, err := hostNetworkSystemFromHostSystem(host)
	if err != nil {
		return networkSystemError(err)
	}
	status.AddToInfo(fmt.Sprintf("Delete portgroup %s on host %s", prefix+strconv.Itoa(int(vlan)), host.Name()))
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	if err := ns.RemovePortGroup(ctx, prefix+strconv.Itoa(int(vlan))); err != nil {
		if err.Error() == itemDoesNotExist {
			status.AddWarning(fmt.Sprintf("Portgroup %s on host %s does not exist", prefix+strconv.Itoa(int(vlan)), host.Name()))
		} else {
			return fmt.Errorf("error deleting port group: %s", err)
		}
	}
	return nil
}

// hostNetworkSystemFromHostSystem locates a HostNetworkSystem from a specified
// HostSystem.
func hostNetworkSystemFromHostSystem(hs *object.HostSystem) (*object.HostNetworkSystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	return hs.ConfigManager().NetworkSystem(ctx)
}

func channelInitialize(numberOfObjects, concurrencyNumber uint) (chan *Networks, *concurrency.Object) {
	in := make(chan *Networks)
	return in, concurrency.New(numberOfObjects, concurrencyNumber)
}

// Loops over the in and ret channels
func channelLooper(in chan *Networks, conObject *concurrency.Object, hosts []*object.HostSystem, vlans *[]uint) error {
	go func() {
		// loop over all items
		for _, e := range *vlans {
			for _, ee := range hosts {
				network := Networks{
					Vlan: e,
					Host: ee,
				}
				in <- &network
			}
		}
		close(in)
	}()
	return conObject.ChannelLooperError()
}

func networkSystemError(err error) error {
	return fmt.Errorf("error loading host network system: %s", err)
}
