package frontend

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/demos"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/ping"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/tasks"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/templates"
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

	myRouter.HandleFunc(pathPrefix+"/access/permissions", ping.Pong).Methods("GET") //returns the users permissions
	myRouter.HandleFunc(pathPrefix+"/demos", ping.Pong).Methods("GET")              //gets the users list of demos
	myRouter.HandleFunc(pathPrefix+"/demos", demos.Post).Methods("POST")            //creates a new demo for the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", ping.Pong).Methods("GET")         //gets information of a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", demos.IdPut).Methods("PUT")       //updates information on a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", demos.IdDelete).Methods("DELETE") //removes a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/login", ping.Pong).Methods("PUT")              //returns a session token
	myRouter.HandleFunc(pathPrefix+"/logout", ping.Pong).Methods("PUT")             //revokes the users session token
	myRouter.HandleFunc(pathPrefix+"/networks", ping.Pong).Methods("POST")          //lists all the networks of vms in a folder and subfolders
	myRouter.HandleFunc(pathPrefix+"/ping", ping.Pong).Methods("GET")               //check if the application is still running
	myRouter.HandleFunc(pathPrefix+"/tasks", tasks.Get).Methods("GET")              //returns the list of all enden,running and queued tasks
	myRouter.HandleFunc(pathPrefix+"/tasks/{id}", tasks.Get).Methods("GET")         //returns the status of the specified task
	myRouter.HandleFunc(pathPrefix+"/templates", templates.Get).Methods("GET")      //returns a list of all availible templates
	myRouter.HandleFunc(pathPrefix+"/templates", ping.Pong).Methods("POST")         //imports a new template from vmware, and returns a task ID
	myRouter.HandleFunc(pathPrefix+"/templates/{id}", templates.Get).Methods("GET") //gets the settings of a template
	myRouter.HandleFunc(pathPrefix+"/templates/{id}", ping.Pong).Methods("PUT")     //updates the settings of a template

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(port)), myRouter))
}
