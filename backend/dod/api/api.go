package frontend

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/demos"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/networks"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/permission"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/ping"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/tasks"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/templates"
	"github.com/Tinyblargon/DemoOnDemand/backend/dod/api/templates/children"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func HandleRequests(pathPrefix string, port uint16, accessLog io.Writer) (err error) {
	router := mux.NewRouter().StrictSlash(true)

	router.Handle(pathPrefix+"/demos", authMiddleware(demos.GetHandler)).Methods("GET")   //gets the users list of demos
	router.Handle(pathPrefix+"/demos", authMiddleware(demos.PostHandler)).Methods("POST") //creates a new demo for the user

	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdGetHandler)).Methods("GET")       //gets information of a specific demo of the user
	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdPutHandler)).Methods("PUT")       //updates information on a specific demo of the user
	router.Handle(pathPrefix+"/demos/{id}", authMiddleware(demos.IdDeleteHandler)).Methods("DELETE") //removes a specific demo of the user

	// TODO remove session token
	// router.HandleFunc(pathPrefix+"/logout", ping.Pong).Methods("PUT") //revokes the users session token
	router.HandleFunc(pathPrefix+"/login", authenticate).Methods("POST") //Authenticates the user/returns a session token

	router.Handle(pathPrefix+"/networks", authMiddleware(networks.PostHandler)).Methods("POST") //lists all the networks of vms in a folder and sub folders

	router.Handle(pathPrefix+"/permission", authMiddleware(permission.GetHandler)).Methods("GET") //returns the users permissions

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

	origins := handlers.AllowedOrigins([]string{"*"})
	headers := handlers.AllowedHeaders([]string{"Authorization", "accept", "Content-Type"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	return http.ListenAndServe(":"+strconv.Itoa(int(port)), handlers.CORS(origins, headers, methods)(handlers.LoggingHandler(accessLog, router)))
}
