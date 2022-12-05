package children

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/filesystem"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/scheduler/job"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/template"
	"github.com/gorilla/mux"
)

type Data struct {
	Children uint `json:"children"`
}

var IdGetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !api.IfRole(r, []string{"root", "admin"}) {
		api.OutputInvalidPermission(w)
		return
	}
	children, err := database.CountTemplateInUse(global.DB, id)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	data := Data{
		Children: children,
	}
	response := api.JsonResponse{
		Data: data,
	}
	response.Output(w)
})

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if !api.IfRole(r, []string{"root"}) {
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
		Config:       &newConfig,
		ChildDestroy: true,
	}
	newJob := job.Job{
		Template: &newTemplate,
	}
	api.NewJob(w, &newJob, r.Header.Get("name"))
})
