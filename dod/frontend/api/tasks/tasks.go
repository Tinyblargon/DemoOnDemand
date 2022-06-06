package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/gorilla/mux"
)

type Data struct {
	Tasks *[]*Task `json:"tasks"`
}

type Task struct {
	ID   uint `json:"id"`
	Info Info `json:"info"`
}

type Info struct {
	User   string `json:"user"`
	Status string `json:"status"`
}

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id != "" {
		status := string(scheduler.Main.GetTaskStatus(id))
		if status == "" {
			status = "Task with id " + id + " does not exist."
		}
		fmt.Fprint(w, status)
	} else {
		allTasks := scheduler.Main.ListAllTasks()
		nuberOfTasks := len(allTasks)
		tasksList := make([]*Task, nuberOfTasks)
		for i, e := range allTasks {
			id, _ := strconv.Atoi(e.ID)
			newTask := new(Task)
			newTask.ID = uint(id)
			newTask.Info.User = e.UserID
			newTask.Info.Status = e.Status.Status
			tasksList[i] = newTask
		}
		data := new(Data)
		data.Tasks = &tasksList
		j, _ := json.Marshal(data)
		fmt.Fprint(w, string(j))
	}
}
