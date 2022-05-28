package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/gorilla/mux"
)

func HandleRequests(pathPrefix string, port uint) {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc(pathPrefix+"/access/permissions", pong).Methods("GET")     //returns the users permissions
	myRouter.HandleFunc(pathPrefix+"/demos", pong).Methods("GET")                  //gets the users list of demos
	myRouter.HandleFunc(pathPrefix+"/demos", pong).Methods("POST")                 //creates a new demo for the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", pong).Methods("GET")             //gets information of a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/demos/{id}", pong).Methods("PUT")             //updates information on a specific demo of the user
	myRouter.HandleFunc(pathPrefix+"/networks", pong).Methods("POST")              //lists all the networks of vms in a folder and subfolders
	myRouter.HandleFunc(pathPrefix+"/ping", pong).Methods("GET")                   //check if the application is still running
	myRouter.HandleFunc(pathPrefix+"/templates", templatesGet).Methods("GET")      //lists all the availible templates
	myRouter.HandleFunc(pathPrefix+"/templates", pong).Methods("POST")             //imports a new template from vmware
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
	listAll := true
	var j []byte
	for _, e := range templates {
		if e == id {
			listAll = false
			var templateConfig *demo.DemoConfig
			templateConfig, err = demo.GetTemplate("demo-01")
			j, err = json.Marshal(templateConfig)
		}
	}
	if listAll {
		j, err = json.Marshal(templates)
	}
	if err == nil {
		fmt.Fprintf(w, string(j))
	}
}
