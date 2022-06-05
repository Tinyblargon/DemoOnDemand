package demos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/gorilla/mux"
)

func Post(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.ReadingBodyFailed(w, err)
		return
	}
	newDemo := job.Demo{
		Create: true,
	}
	err = json.Unmarshal(body, &newDemo)
	if err != nil {
		api.ReadingBodyFailed(w, err)
		return
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob)
}

func IdDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	demoName := strings.Split(id, "_")
	if len(demoName) < 3 {
		fmt.Fprintf(w, api.InvalidID)
	}
	demoNumber, err := strconv.Atoi(demoName[2])
	if err != nil {
		fmt.Fprintf(w, api.InvalidID)
	}
	newDemo := job.Demo{
		Template: demoName[1],
		UserName: demoName[0],
		Number:   uint(demoNumber),
		Destroy:  true,
	}
	newjob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newjob)
}
