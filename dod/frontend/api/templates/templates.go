package templates

import (
	"net/http"

	demoactions "github.com/Tinyblargon/DemoOnDemand/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/filesystem"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/template"
	"github.com/gorilla/mux"
)

type Data struct {
	TemplateConfig *template.Config `json:"config,omitempty"`
	Templates      *[]string        `json:"templates,omitempty"`
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
	var templateConfig *template.Config
	var templateList *[]string
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	if id != "" {
		if !util.IsStringUnique(&templates, id) {
			templateConfig, err = template.Get(id)
			if err != nil {
				api.OutputServerError(w, "", err)
				return
			}
		}
	} else {
		templateList = &templates
	}

	data := Data{
		TemplateConfig: templateConfig,
		Templates:      templateList,
	}
	response := api.JsonResponse{
		Data: data,
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
	c, err := session.New(*global.VMwareConfig)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	networks, err := demoactions.GetImportProperties(c.VimClient, datacenter.GetObject(), newConfig.Path)
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
	newjob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newjob, r.Header.Get("name"))
}

func IdDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !api.IfRole(r, []string{"admin", "root"}) {
		api.OutputInvalidPermission(w)
		return
	}
	if !filesystem.CheckExistance(global.ConfigFolder + "/" + id) {
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
	newjob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newjob, r.Header.Get("name"))
}
