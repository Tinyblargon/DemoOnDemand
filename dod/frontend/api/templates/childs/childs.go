package childs

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/file"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/template"
	"github.com/gorilla/mux"
)

type Data struct {
	Childs uint `json:"childs"`
}

var IdGetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdGet(w, r)
})

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdDelete(w, r)
})

func IdGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !api.IfRole(r, []string{"root", "admin"}) {
		api.OutputInvalidPermission(w)
		return
	}
	childs, err := database.CountTemplateInUse(global.DB, id)
	if err != nil {
		api.OutputServerError(w, "")
		// TODO
		// LOG this error to file
		return
	}
	data := Data{
		Childs: childs,
	}
	response := api.JsonResponse{
		Data: data,
	}
	response.Output(w)
}

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
