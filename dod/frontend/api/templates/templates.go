package templates

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/util"
	"github.com/gorilla/mux"
)

type Data struct {
	TemplateConfig *demo.DemoConfig `json:"config,omitempty"`
	Templates      *[]string        `json:"templates,omitempty"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	templates, err := demo.ListTemplates()
	var templateConfig *demo.DemoConfig
	var templateList *[]string
	if err != nil {
		// TODO
		// Log the error to error log
	}
	if id != "" {
		if !util.IsStringUnique(&templates, id) {
			templateConfig, err = demo.GetTemplate(id)
			if err != nil {
				// TODO
				// Log the error to error log
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
