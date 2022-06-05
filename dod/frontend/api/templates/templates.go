package templates

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/gorilla/mux"
)

func Get(w http.ResponseWriter, r *http.Request) {
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
