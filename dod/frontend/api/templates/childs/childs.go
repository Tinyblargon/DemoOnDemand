package childs

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/file"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/template"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/gorilla/mux"
)

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdDelete(w, r)
})

func IdDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !api.IfRole(r, []string{"root"}) {
		api.OutputInvalidPermission(w)
		return
	}
	if !file.CheckExistance(global.ConfigFolder + "/" + id) {
		api.OutputInvalidID(w)
		return
	}
	newConfig := template.Config{
		Name: id,
	}
	newTemplate := job.Template{
		Config:       newConfig,
		ChildDestroy: true,
	}
	newjob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newjob, r.Header.Get("name"))
}
