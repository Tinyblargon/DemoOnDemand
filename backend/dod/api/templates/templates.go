package templates

import (
	"context"
	"net/http"

	demoactions "github.com/Tinyblargon/DemoOnDemand/backend/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/filesystem"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/vsphere/session"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/scheduler/job"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/template"
	"github.com/gorilla/mux"
)

type Data struct {
	TemplateConfig *template.Config   `json:"config,omitempty"`
	Templates      *[]template.Config `json:"templates,omitempty"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
})

var PostHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if api.ErrorToManyNetworks(w, &networks) {
		return
	}
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
})

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
})
