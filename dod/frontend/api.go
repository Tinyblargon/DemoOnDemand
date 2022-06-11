package frontend

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/demos"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/networks"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/permissions"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/ping"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/tasks"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/templates"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Post struct {
	Userid string `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func HandleRequests(pathPrefix string, port uint) {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc(pathPrefix+"/auth", authenticate).Methods("POST") //Authenticates the user/returns a session token

	router.Handle(pathPrefix+"/demos", authMiddleware(demos.GetHandler)).Methods("GET")   //gets the users list of demos
	router.Handle(pathPrefix+"/demos", authMiddleware(demos.PostHandler)).Methods("POST") //creates a new demo for the user

	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdGetHandler)).Methods("GET")       //gets information of a specific demo of the user
	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdPutHandler)).Methods("PUT")       //updates information on a specific demo of the user
	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdDeleteHandler)).Methods("DELETE") //removes a specific demo of the user

	router.HandleFunc(pathPrefix+"/logout", ping.Pong).Methods("PUT") //revokes the users session token

	router.Handle(pathPrefix+"/networks", authMiddleware(networks.PostHandler)).Methods("POST") //lists all the networks of vms in a folder and subfolders

	router.Handle(pathPrefix+"/permissions", authMiddleware(permissions.GetHandler)).Methods("GET") //returns the users permissions

	router.HandleFunc(pathPrefix+"/ping", ping.Pong).Methods("GET") //check if the application is still running

	router.HandleFunc(pathPrefix+"/renew", ping.Pong).Methods("POST") //Returns a new session token. This is used to extend the session.

	router.Handle(pathPrefix+"/tasks", authMiddleware(tasks.GetHandler)).Methods("GET") //returns the list of all enden,running and queued tasks

	router.Handle(pathPrefix+"/tasks/{id}", authMiddleware(tasks.GetHandler)).Methods("GET") //returns the status of the specified task

	router.Handle(pathPrefix+"/templates", authMiddleware(templates.GetHandler)).Methods("GET") //returns a list of all availible templates
	router.HandleFunc(pathPrefix+"/templates", ping.Pong).Methods("POST")                       //imports a new template from vmware, and returns a task ID

	router.Handle(pathPrefix+"/templates/{id}", authMiddleware(templates.GetHandler)).Methods("GET") //gets the settings of a template
	router.HandleFunc(pathPrefix+"/templates/{id}", ping.Pong).Methods("PUT")                        //updates the settings of a template
	router.HandleFunc(pathPrefix+"/templates/{id}", ping.Pong).Methods("DELETE")                     //deletes a template

	// TODO
	// This should log to a file instead of os.Stdout
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(port)), handlers.LoggingHandler(os.Stdout, router)))
}
