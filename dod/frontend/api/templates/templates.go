package templates

import (
	"context"
	"net/http"

	demoactions "github.com/Tinyblargon/DemoOnDemand/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/filesystem"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/template"
	"github.com/gorilla/mux"
)

type Data struct {
	TemplateConfig *template.Config   `json:"config,omitempty"`
	Templates      *[]template.Config `json:"templates,omitempty"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

var PostHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Post(w, r)
})

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdDelete(w, r)
})

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	templates, err := template.List()
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	var templateConfigList *[]template.Config
	var templateConfig *template.Config
	if id != "" {
		if !util.IsStringUnique(&templates, id) {
			templateConfig, err = template.Get(id)
			if err != nil {
				api.OutputServerError(w, "", err)
				return
			}
		}
	} else {
		templateConfigList, err = template.GetDescriptions(&templates)
		if err != nil {
			api.OutputServerError(w, "", err)
			return
		}
	}
	response := api.JsonResponse{
		Data: Data{
			TemplateConfig: templateConfig,
			Templates:      templateConfigList,
		},
	}
	response.Output(w)
}

func Post(w http.ResponseWriter, r *http.Request) {
	newConfig := template.Config{}
	err := api.GetBody(r, &newConfig)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}
	if !api.IfRole(r, []string{"admin", "root"}) {
		api.OutputInvalidPermission(w)
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
	networks, err := demoactions.GetImportProperties(c.VimClient, dataCenter, newConfig.Path)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	api.ErrorToManyNetworks(w, &networks)
	newConfig.Defaults()
	err = newConfig.Validate(false)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}
	newTemplate := job.Template{
		Config: &newConfig,
		Import: true,
	}
	newJob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newJob, r.Header.Get("name"))
}

func IdDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !api.IfRole(r, []string{"admin", "root"}) {
		api.OutputInvalidPermission(w)
		return
	}
	if !filesystem.CheckExistence(global.ConfigFolder + "/" + id) {
		api.OutputInvalidID(w)
		return
	}
	newConfig := template.Config{
		Name: id,
	}
	newTemplate := job.Template{
		Config:  &newConfig,
		Destroy: true,
	}
	newJob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newJob, r.Header.Get("name"))
}
