package tasks

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler"
	"github.com/gorilla/mux"
)

type Data struct {
	Tasks *[]*scheduler.Task `json:"tasks"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	name := r.Header.Get("name")
	role := r.Header.Get("role")
	if id != "" {
		taskID, err := strconv.Atoi(id)
		if err != nil || taskID <= 0 {
			api.OutputUserInputError(w, "id should be a positive number")
			return
		}
		info, userID := scheduler.Main.GetTaskStatus(uint(taskID))
		infoString := string(info)
		if infoString != "" {
			if !api.IfRoleOrUser(r, "root", userID) {
				api.OutputInvalidPermission(w)
				return
			}
		} else {
			infoString = "Task with id " + id + " does not exist."
		}
		fmt.Fprint(w, infoString)
	} else {
		allTasks := scheduler.Main.ListAllTasks()
		var tasksList []*scheduler.Task
		if role == "root" {
			tasksList = allTasks
		} else {
			tasksList = make([]*scheduler.Task, 0)
			for _, e := range allTasks {
				if e.Info.UserID == name {
					tasksList = append(tasksList, e)
				}
			}
		}
		data := new(Data)
		data.Tasks = &tasksList
		response := api.JsonResponse{
			Data: &data,
		}
		response.Output(w)
	}
}
