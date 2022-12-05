package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

func NewFinder(client *govmomi.Client, DataCenter *object.Datacenter, Path string) (context.Context, context.CancelFunc, *find.Finder, string) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	return ctx, cancel, find.NewFinder(client.Client, false), "/" + DataCenter.Name() + "/vm/" + strings.Trim(Path, "/")
}

func RunTaskWait(task *object.Task, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	if err := task.Wait(ctx); err != nil {
		return fmt.Errorf("error on waiting for '%s' task completion: %s", message, err)
	}
	return nil
}
