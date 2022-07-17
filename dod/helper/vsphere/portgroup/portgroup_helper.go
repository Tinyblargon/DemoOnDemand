package portgroup

import (
	"fmt"
	"strconv"

	"context"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/concurrency"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/taskstatus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const hostPortGroupIDPrefix = "tf-HostPortGroup"

type Networks struct {
	Vlan uint
	Host *object.HostSystem
}

func Create(c *govmomi.Client, hosts []*object.HostSystem, vlans *[]uint, prefix, vSwitch string, concurrencyNumber uint, status *taskstatus.Status) (err error) {
	numberOfTasks := uint(len(*vlans) * len(hosts))
	in, ret, concurrencyNumber := channelInitialize(numberOfTasks, concurrencyNumber)
	// spawn "concurrencyNumber" amount of threads
	for x := 0; x < int(concurrencyNumber); x++ {
		go func() {
			for x := range in {
				ret <- createSingle(c, x.Host, prefix, vSwitch, x.Vlan, status)
			}
		}()
	}
	err = channelLooper(in, ret, hosts, vlans, numberOfTasks)
	return
}

// Creates a portgroup on a singular host
func createSingle(c *govmomi.Client, host *object.HostSystem, prefix, vSwitch string, vlan uint, status *taskstatus.Status) (err error) {
	ns, err := hostNetworkSystemFromHostSystem(host)
	spec := expandHostPortGroupSpec(prefix, vSwitch, vlan)
	status.AddToInfo(fmt.Sprintf("[DEBUG] Create portgroup %s on host %s", spec.Name, host.Name()))
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
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
	obj := &types.HostPortGroupSpec{
		Name:        prefix + strconv.Itoa(int(vlan)),
		VlanId:      int32(vlan),
		VswitchName: vSwitchName,
		// Policy:      *expandHostNetworkPolicy(d),
	}
	return obj
}

// ##############################################################################################

func Delete(c *govmomi.Client, hosts []*object.HostSystem, prefix string, vlans *[]uint, concurrencyNumber uint, status *taskstatus.Status) (err error) {
	numberOfTasks := uint(len(*vlans) * len(hosts))
	in, ret, concurrencyNumber := channelInitialize(numberOfTasks, concurrencyNumber)
	// spawn "concurrencyNumber" amount of threads
	for x := 0; x < int(concurrencyNumber); x++ {
		go func() {
			for x := range in {
				ret <- deleteSingle(c, x.Host, prefix, x.Vlan, status)
			}
		}()
	}
	return channelLooper(in, ret, hosts, vlans, numberOfTasks)
}

// Deletest a portgroup on a singular host
func deleteSingle(c *govmomi.Client, host *object.HostSystem, prefix string, vlan uint, status *taskstatus.Status) error {
	ns, err := hostNetworkSystemFromHostSystem(host)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}
	status.AddToInfo(fmt.Sprintf("[DEBUG] Delete portgroup %s on host %s", prefix+strconv.Itoa(int(vlan)), host.Name()))
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	if err := ns.RemovePortGroup(ctx, prefix+strconv.Itoa(int(vlan))); err != nil {
		return fmt.Errorf("error deleting port group: %s", err)
	}
	return nil
}

// hostNetworkSystemFromHostSystem locates a HostNetworkSystem from a specified
// HostSystem.
func hostNetworkSystemFromHostSystem(hs *object.HostSystem) (*object.HostNetworkSystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return hs.ConfigManager().NetworkSystem(ctx)
}

func channelInitialize(numberOfObjects, concurrencyNumner uint) (chan *Networks, chan error, uint) {
	in := make(chan *Networks)
	ret := make(chan error)
	return in, ret, concurrency.DecideMinimumTreads(numberOfObjects, concurrencyNumner)
}

// Loops over the in and ret channels
func channelLooper(in chan *Networks, ret chan error, hosts []*object.HostSystem, vlans *[]uint, cycles uint) (err error) {
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
	err = concurrency.ChannelLooperError(ret, cycles)
	close(ret)
	return
}
