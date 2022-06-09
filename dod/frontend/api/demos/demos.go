package demos

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/gorilla/mux"
)

type StartStopRestart struct {
	Task string `json:"task"`
}

type Data struct {
	Demos *[]*database.Demo `json:"demos"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

var PostHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Post(w, r)
})

var IdGetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdGet(w, r)
})

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdDelete(w, r)
})

var IdPutHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdPut(w, r)
})

func Get(w http.ResponseWriter, r *http.Request) {
	var demos *[]*database.Demo
	if r.Header.Get("role") == "root" {
		// demos = new(database.Demos)
		var err error
		demos, err = database.ListAllDemos(global.DB)
		if err != nil {
			// TODO
			// Log error to file
			return
		}
	} else {
		var err error
		demos, err = database.ListDemosOfUser(global.DB, r.Header.Get("name"))
		if err != nil {
			// TODO
			// Log error to file
			return
		}
	}
	data := Data{
		Demos: demos,
	}
	response := api.JsonResponse{
		Data: data,
	}
	response.Output(w)
}

func Post(w http.ResponseWriter, r *http.Request) {
	newDemo := job.Demo{
		Create: true,
	}
	err := api.GetBody(w, r, &newDemo)
	if err != nil {
		return
	}
	if !api.IfRoleOrUser(r, "root", newDemo.UserName) {
		api.OutputInvalidPermission(w)
		return
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob, newDemo.UserName)
}

type IdData struct {
	Demo *database.Demo `json:"demo"`
}

func IdGet(w http.ResponseWriter, r *http.Request) {
	userName, demoName, demoNumber := checkID(w, r)
	if !api.IfRoleOrUser(r, "root", userName) {
		api.OutputInvalidPermission(w)
		return
	}
	demo, err := database.GetSpecificDemo(global.DB, userName, demoName, uint(demoNumber))
	if err != nil {
		return
	}
	data := IdData{
		Demo: demo,
	}
	response := api.JsonResponse{
		Data: data,
	}
	response.Output(w)
}

func IdDelete(w http.ResponseWriter, r *http.Request) {
	userName, demoName, demoNumber := checkID(w, r)
	if !api.IfRoleOrUser(r, "root", userName) {
		api.OutputInvalidPermission(w)
		return
	}
	newDemo := job.Demo{
		Template: demoName,
		UserName: userName,
		Number:   uint(demoNumber),
		Destroy:  true,
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob, userName)
}

func IdPut(w http.ResponseWriter, r *http.Request) {
	userName, demoName, demoNumber := checkID(w, r)
	if !api.IfRoleOrUser(r, "root", userName) {
		api.OutputInvalidPermission(w)
		return
	}
	SSR := StartStopRestart{}
	err := api.GetBody(w, r, &SSR)
	if err != nil {
		return
	}
	newDemo := job.Demo{
		Template: demoName,
		UserName: userName,
		Number:   uint(demoNumber),
	}
	switch SSR.Task {
	case "start":
		newDemo.Start = true
	case "stop":
		newDemo.Stop = true
	case "restart":
		newDemo.Start = true
		newDemo.Stop = true
	default:
		fmt.Fprint(w, "Key task shoud be (Start|Stop|Restart)")
		return
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob, userName)
}

func checkID(w http.ResponseWriter, r *http.Request) (username, demoName string, demoNumber int) {
	vars := mux.Vars(r)
	id := vars["id"]
	demoString := strings.Split(id, "_")
	if len(demoString) != 3 {
		fmt.Fprintf(w, api.InvalidID)
	}
	demoNumber, err := strconv.Atoi(demoString[2])
	if err != nil {
		fmt.Fprintf(w, api.InvalidID)
	}
	username = demoString[0]
	demoName = demoString[1]
	return
}
