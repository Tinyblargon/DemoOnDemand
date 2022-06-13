package templates

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
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

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	templates, err := template.List()
	var templateConfig *template.Config
	var templateList *[]string
	if err != nil {
		// TODO
		// Log the error to error log
		return
	}
	if id != "" {
		if !util.IsStringUnique(&templates, id) {
			templateConfig, err = template.Get(id)
			if err != nil {
				// TODO
				// Log the error to error log
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
	newConfig.Defaults()
	err = newConfig.Validate(false)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}
	newTemplate := job.Template{
		Config: newConfig,
		Import: true,
	}
	newjob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newjob, r.Header.Get("name"))
}
