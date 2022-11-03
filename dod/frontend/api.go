package frontend

import (
	"net/http"
	"os"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/demos"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/networks"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/permissions"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/ping"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/tasks"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/templates"
	"github.com/Tinyblargon/DemoOnDemand/dod/frontend/api/templates/children"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Post struct {
	UserID string `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func HandleRequests(logFile, pathPrefix string, port uint16) (err error) {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc(pathPrefix+"/auth", authenticate).Methods("POST") //Authenticates the user/returns a session token

	router.Handle(pathPrefix+"/demos", authMiddleware(demos.GetHandler)).Methods("GET")   //gets the users list of demos
	router.Handle(pathPrefix+"/demos", authMiddleware(demos.PostHandler)).Methods("POST") //creates a new demo for the user

	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdGetHandler)).Methods("GET")       //gets information of a specific demo of the user
	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdPutHandler)).Methods("PUT")       //updates information on a specific demo of the user
	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdDeleteHandler)).Methods("DELETE") //removes a specific demo of the user

	// TODO remove session token
	// router.HandleFunc(pathPrefix+"/logout", ping.Pong).Methods("PUT") //revokes the users session token

	router.Handle(pathPrefix+"/networks", authMiddleware(networks.PostHandler)).Methods("POST") //lists all the networks of vms in a folder and sub folders

	router.Handle(pathPrefix+"/permissions", authMiddleware(permissions.GetHandler)).Methods("GET") //returns the users permissions

	router.HandleFunc(pathPrefix+"/ping", ping.Pong).Methods("GET") //check if the application is still running

	// TODO refresh session token
	// router.HandleFunc(pathPrefix+"/renew", ping.Pong).Methods("POST") //Returns a new session token. This is used to extend the session.

	router.Handle(pathPrefix+"/tasks", authMiddleware(tasks.GetHandler)).Methods("GET") //returns the list of all ended,running and queued tasks

	router.Handle(pathPrefix+"/tasks/{id}", authMiddleware(tasks.GetHandler)).Methods("GET") //returns the status of the specified task

	router.Handle(pathPrefix+"/templates", authMiddleware(templates.GetHandler)).Methods("GET")   //returns a list of all available templates
	router.Handle(pathPrefix+"/templates", authMiddleware(templates.PostHandler)).Methods("POST") //imports a new template from vmware, and returns a task ID

	router.Handle(pathPrefix+"/templates/{id}", authMiddleware(templates.GetHandler)).Methods("GET")         //gets the settings of a template
	router.HandleFunc(pathPrefix+"/templates/{id}", ping.Pong).Methods("PUT")                                //updates the settings of a template
	router.Handle(pathPrefix+"/templates/{id}", authMiddleware(templates.IdDeleteHandler)).Methods("DELETE") //deletes a template

	router.Handle(pathPrefix+"/templates/{id}/children", authMiddleware(children.IdGetHandler)).Methods("GET")       //returns the amount of demos that exist based on the speciefied template
	router.Handle(pathPrefix+"/templates/{id}/children", authMiddleware(children.IdDeleteHandler)).Methods("DELETE") //deletes all demos based on the specified template

	accessLog, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	return http.ListenAndServe(":"+strconv.Itoa(int(port)), handlers.LoggingHandler(accessLog, router))
}
