package demos

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/gorilla/mux"
)

type StartStopRestart struct {
	Task string `json:"task"`
}

func Post(w http.ResponseWriter, r *http.Request) {
	newDemo := job.Demo{
		Create: true,
	}
	err := api.GetBody(w, r, &newDemo)
	if err != nil {
		return
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob, "placeholder")
}

func IdDelete(w http.ResponseWriter, r *http.Request) {
	demoName, demoNumber := checkID(w, r)
	newDemo := job.Demo{
		Template: demoName[1],
		UserName: demoName[0],
		Number:   uint(demoNumber),
		Destroy:  true,
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob, "placeholder")
}

func IdPut(w http.ResponseWriter, r *http.Request) {
	demoName, demoNumber := checkID(w, r)
	SSR := StartStopRestart{}
	err := api.GetBody(w, r, &SSR)
	if err != nil {
		return
	}
	newDemo := job.Demo{
		Template: demoName[1],
		UserName: demoName[0],
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
	api.NewJob(w, &newjob, "placeholder")
}

func checkID(w http.ResponseWriter, r *http.Request) (demoName []string, demoNumber int) {
	vars := mux.Vars(r)
	id := vars["id"]
	demoName = strings.Split(id, "_")
	if len(demoName) < 3 {
		fmt.Fprintf(w, api.InvalidID)
	}
	demoNumber, err := strconv.Atoi(demoName[2])
	if err != nil {
		fmt.Fprintf(w, api.InvalidID)
	}
	return
}
