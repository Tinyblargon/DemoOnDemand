package permissions

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
)

type JsonResponse struct {
	Data *Data `json:"data"`
}

type Data struct {
	User string `json:"user,omitempty"`
	Role string `json:"role,omitempty"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

func Get(w http.ResponseWriter, r *http.Request) {
	data := Data{
		User: r.Header.Get("name"),
		Role: r.Header.Get("role"),
	}
	jsonResponse := JsonResponse{
		Data: &data,
	}
	api.OutputJson(w, jsonResponse)
}
