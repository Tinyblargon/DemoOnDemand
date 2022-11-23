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

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

// TODO return start and end date/time of task

func Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	name := r.Header.Get("name")
	role := r.Header.Get("role")
	if id != "" {
		info, userID := scheduler.Main.GetTaskStatus(id)
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
		var tasksList []*Task
		if role == "root" {
			numberOfTasks := len(allTasks)
			tasksList = make([]*Task, numberOfTasks)
			for i, e := range allTasks {
				newTask := newTask(e)
				tasksList[i] = newTask
			}
		} else {
			tasksList = make([]*Task, 0)
			for _, e := range allTasks {
				if e.UserID == name {
					newTask := newTask(e)
					tasksList = append(tasksList, newTask)
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

func newTask(task *scheduler.Task) (newTask *Task) {
	id, _ := strconv.Atoi(task.ID)
	newTask = new(Task)
	newTask.ID = uint(id)
	newTask.Info.User = task.UserID
	newTask.Info.Status = task.Status.Status
	return
}
