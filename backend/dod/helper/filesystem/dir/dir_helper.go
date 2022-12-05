package dir

import (
	"os"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/helper/filesystem"
)

func Create(path string) (err error) {
	if !filesystem.CheckExistence(path) {
		err = os.Mkdir(path, 0755)
	}
	return
}
