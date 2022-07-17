package demo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
)

const invalidString string = "invalid ID"

type Demo struct {
	Name string
	User string
	ID   uint
}

func (demoObj Demo) CreateDemoURl() string {
	return global.DemoFodler + "/" + demoObj.CreateID()
}

// Creates a demo ID from its seperate parts
func (demoObj Demo) CreateID() string {
	return demoObj.User + "_" + strconv.Itoa(int(demoObj.ID)) + "_" + demoObj.Name
}

// Spits the demo ID into its seperate parts
func ReverseID(id string) (username, demoName string, demoNumber uint, err error) {
	idString := strings.Split(id, "_")
	if len(idString) != 3 {
		err = fmt.Errorf(invalidString)
		return
	}
	tmpNumber, err := strconv.Atoi(idString[1])
	if err != nil {
		err = fmt.Errorf(invalidString)
		return
	}
	demoNumber = uint(tmpNumber)
	username = idString[0]
	demoName = idString[2]
	return
}
