package generic

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

func NewFinder(client *govmomi.Client, DataCenter, Path string) (context.Context, context.CancelFunc, *find.Finder, string) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	return ctx, cancel, find.NewFinder(client.Client, false), "/" + DataCenter + "/vm/" + strings.Trim(Path, "/")
}

func RunTaskWait(task *object.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	if err := task.Wait(ctx); err != nil {
		return fmt.Errorf("error on waiting for deletion task completion: %s", err)
	}
	return nil
}
