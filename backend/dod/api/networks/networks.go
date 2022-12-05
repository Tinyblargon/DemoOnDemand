package networks

import (
	"context"
	"net/http"

	demoactions "github.com/Tinyblargon/DemoOnDemand/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/folder"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/session"
)

type Input struct {
	Path string `json:"path,omitempty"`
}

type Data struct {
	Networks []string `json:"networks,omitempty"`
}

var PostHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if !api.IfRole(r, []string{"root", "admin"}) {
		api.OutputInvalidPermission(w)
		return
	}
	input := Input{}
	err := api.GetBody(r, &input)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}
	c, err := session.New(*vsphere.GetConfig())
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	defer c.VimClient.Logout(ctx)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	dataCenter, err := datacenter.Get(c.VimClient, datacenter.GetName())
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	if !folder.Exists(c.VimClient, dataCenter, folder.VSphereFolderTypeVM, input.Path) {
		api.OutputUserInputError(w, "folder does not exist")
		return
	}
	networks, err := demoactions.GetImportProperties(c.VimClient, dataCenter, input.Path)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	if api.ErrorToManyNetworks(w, &networks) {
		return
	}
	data := Data{
		Networks: networks,
	}
	response := api.JsonResponse{
		Data: &data,
	}
	response.Output(w)
})
