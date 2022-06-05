package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends"
	"github.com/gorilla/mux"
)

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id != "" {
		status := string(backends.Main.GetTaskStatus(id))
		if status == "" {
			status = "Task with id " + id + " does not exist."
		}
		fmt.Fprintf(w, status)
	} else {
		j, _ := json.Marshal(backends.Main.ListAllTasks())
		fmt.Fprintf(w, string(j))
	}
}
