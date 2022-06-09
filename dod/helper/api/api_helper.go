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

type JsonResponse struct {
	Data any `json:"data"`
}

func (j *JsonResponse) Output(w http.ResponseWriter) {
	response, _ := json.Marshal(j)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(response))
}

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

func NewJob(w http.ResponseWriter, newJob *job.Job, jobOwner string) {
	fmt.Fprintf(w, "Task added with ID: %s", scheduler.Main.Add(newJob, 9999999, jobOwner))
}

func OutputInvalidPermission(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprint(w, InvalidPerm)
}

func IfRoleOrUser(r *http.Request, role, user string) bool {
	if r.Header.Get("role") != role && r.Header.Get("name") != user {
		return false
	}
	return true
}
