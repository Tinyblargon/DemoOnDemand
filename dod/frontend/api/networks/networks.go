package networks

import (
	"context"
	"net/http"

	demoactions "github.com/Tinyblargon/DemoOnDemand/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
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
	Post(w, r)
})

func Post(w http.ResponseWriter, r *http.Request) {
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
	networks, err := demoactions.GetImportProperties(c.VimClient, datacenter.GetObject(), input.Path)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	api.ErrorToManyNetworks(w, &networks)
	data := Data{
		Networks: networks,
	}
	response := api.JsonResponse{
		Data: &data,
	}
	response.Output(w)
}
