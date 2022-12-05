package ping

import (
	"fmt"
	"net/http"
)

func Pong(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}
