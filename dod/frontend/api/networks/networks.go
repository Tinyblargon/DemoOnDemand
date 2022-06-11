package networks

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/session"
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
	err := api.GetBody(w, r, &input)
	if err != nil {
		return
	}
	c, err := session.New(*global.VMwareConfig)
	if err != nil {
		return
	}
	networks, err := demo.GetImportProperties(c.VimClient, global.VMwareConfig.DataCenter, input.Path)
	if err != nil {
		return
	}
	data := Data{
		Networks: networks,
	}
	response := api.JsonResponse{
		Data: &data,
	}
	response.Output(w)
}
