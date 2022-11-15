package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/logger"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
)

const InvalidPerm string = "Invalid Permission."
const InvalidID string = "Invalid ID."
const DemoAlreadyExists string = "Demo Already Exists."
const DemoDoesNotExists string = "Demo Does Not Exist."

type JsonResponse struct {
	Data any `json:"data"`
}

func (j *JsonResponse) Output(w http.ResponseWriter) {
	response, _ := json.Marshal(j)
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, string(response))
}

func GetBody(r *http.Request, v any) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading body failed: %s", err)
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("invalid json: %s", err)
	}
	return nil
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

func OutputInvalidID(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, InvalidID)
}

func OutputUserInputError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, err)
}

func OutputServerError(w http.ResponseWriter, message string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if message == "" {
		message = "internal server error"
	}
	fmt.Fprint(w, message)
	logger.Error(err)
}

func OutputDemoAlreadyExists(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, DemoAlreadyExists)
}

func OutputDemoDoesNotExists(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, DemoDoesNotExists)
}

func ErrorToManyNetworks(w http.ResponseWriter, networks *[]string) {
	if len(*networks) > 8 {
		OutputUserInputError(w, fmt.Errorf("to many networks found, found %d, maximum is 8", len(*networks)).Error())
		return
	}
}

func IfRoleOrUser(r *http.Request, role, user string) bool {
	if r.Header.Get("role") != role && r.Header.Get("name") != user {
		return false
	}
	return true
}

func IfRole(r *http.Request, roles []string) bool {
	role := r.Header.Get("role")
	for _, e := range roles {
		if e == role {
			return true
		}
	}
	return false
}
