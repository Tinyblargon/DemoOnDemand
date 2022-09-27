package dir

import (
	"os"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/filesystem"
)

func Create(path string) (err error) {
	if !filesystem.CheckExistance(path) {
		err = os.Mkdir(path, 0755)
	}
	return
}
