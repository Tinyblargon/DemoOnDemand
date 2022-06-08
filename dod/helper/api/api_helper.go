package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
)

const InvalidPerm string = "Invalid Permission."

const InvalidID string = "Invalid ID."

func GetBody(w http.ResponseWriter, r *http.Request, v any) (err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Reading body failed: %s", err)
		return
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		fmt.Fprintf(w, "Invalid json: %s", err)
	}
	return
}

func ReadingBodyFailed(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Reading body failed: %s", err)
}

func NewJob(w http.ResponseWriter, newJob *job.Job, userID string) {
	fmt.Fprintf(w, "Task added with ID: %s", scheduler.Main.Add(newJob, 9999999, userID))
}

func OutputJson(w http.ResponseWriter, jsonResponse any) {
	j, _ := json.Marshal(jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(j))
}
