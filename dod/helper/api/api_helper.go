package api

import (
	"fmt"
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/backends"
	"github.com/Tinyblargon/DemoOnDemand/dod/backends/job"
)

const InvalidID string = "Invalid ID."

func ReadingBodyFailed(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Reading body failed: %s", err)
}

func NewJob(w http.ResponseWriter, newJob *job.Job) {
	fmt.Fprintf(w, "Task added with ID %s", backends.Main.Add(newJob, 9999999))
}
