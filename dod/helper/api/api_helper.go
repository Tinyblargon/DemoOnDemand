package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/gorilla/mux"
)

type Post struct {
	Userid string `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func HandleRequests(pathPrefix string, port uint) {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc(pathPrefix+"/access/permissions", pong).Methods("GET")     //returns the users permissions
	myRouter.HandleFunc(pathPrefix+"/demos", pong).Methods("GET")                  //gets the users list of demos
	myRouter.HandleFunc(pathPrefix+"/demos", demosPost).Methods("POST")            //creates a new demo for the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", pong).Methods("GET")             //gets information of a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", pong).Methods("PUT")             //updates information on a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/login", pong).Methods("PUT")                  //returns a session token
	myRouter.HandleFunc(pathPrefix+"/logout", pong).Methods("PUT")                 //revokes the users session token
	myRouter.HandleFunc(pathPrefix+"/networks", pong).Methods("POST")              //lists all the networks of vms in a folder and subfolders
	myRouter.HandleFunc(pathPrefix+"/ping", pong).Methods("GET")                   //check if the application is still running
	myRouter.HandleFunc(pathPrefix+"/tasks", tasksGet).Methods("GET")              //returns the list of all enden,running and queued tasks
	myRouter.HandleFunc(pathPrefix+"/tasks/{id}", tasksGet).Methods("GET")         //returns the status of the specified task
	myRouter.HandleFunc(pathPrefix+"/templates", templatesGet).Methods("GET")      //returns a list of all availible templates
	myRouter.HandleFunc(pathPrefix+"/templates", pong).Methods("POST")             //imports a new template from vmware, and returns a task ID
	myRouter.HandleFunc(pathPrefix+"/templates/{id}", templatesGet).Methods("GET") //gets the settings of a template
	myRouter.HandleFunc(pathPrefix+"/templates/{id}", pong).Methods("PUT")         //updates the settings of a template

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(port)), myRouter))
}

func pong(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func templatesGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	templates, err := demo.ListTemplates()
	var j []byte
	if id != "" {
		for _, e := range templates {
			if e == id {
				var templateConfig *demo.DemoConfig
				templateConfig, err = demo.GetTemplate(e)
				j, err = json.Marshal(templateConfig)
			}
		}
	} else {
		j, err = json.Marshal(templates)
	}
	if err == nil {
		fmt.Fprintf(w, string(j))
	}
}

func tasksGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id != "" {
		status := string(global.Backend.GetTaskStatus(id))
		if status == "" {
			status = "Task with id " + id + " does not exist."
		}
		fmt.Fprintf(w, status)
	} else {
		j, _ := json.Marshal(global.Backend.ListAllTasks())
		fmt.Fprintf(w, string(j))
	}
}

func demosPost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	post := Post{}
	err = json.Unmarshal(body, &post)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
	}
	taskID := global.Backend.Add([]byte(body), 9999999)
	log.Printf("Post added with ID %s", taskID)

}
