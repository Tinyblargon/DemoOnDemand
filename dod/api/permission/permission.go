package permission

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
)

type Data struct {
	User string `json:"user,omitempty"`
	Role string `json:"role,omitempty"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	data := Data{
		User: r.Header.Get("name"),
		Role: r.Header.Get("role"),
	}
	response := api.JsonResponse{
		Data: &data,
	}
	response.Output(w)
})
