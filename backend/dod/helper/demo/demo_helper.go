package demo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tinyblargon/DemoOnDemand/backend/dod/global"
)

const invalidString string = "invalid ID"

type Demo struct {
	Name string
	User string
	ID   uint
}

func (demoObj Demo) CreateDemoURl() string {
	return global.DemoFolder + "/" + demoObj.CreateID()
}

// Creates a demo ID from its separate parts
func (demoObj Demo) CreateID() string {
	return demoObj.User + "_" + strconv.Itoa(int(demoObj.ID)) + "_" + demoObj.Name
}

// creates a demo object from its separate parts
func CreateObject(id string) (demo Demo, err error) {
	idString := strings.Split(id, "_")
	var demoName string
	if len(idString) < 3 {
		err = fmt.Errorf(invalidString)
		return
	}
	if len(idString) > 3 {
		demoName = idString[2]
		for i := 3; i < len(idString); i++ {
			demoName = demoName + "_" + idString[i]
		}
	} else {
		demoName = idString[2]
	}
	tmpNumber, err := strconv.Atoi(idString[1])
	if err != nil {
		err = fmt.Errorf(invalidString)
		return
	}
	demo = Demo{
		Name: demoName,
		User: idString[0],
		ID:   uint(tmpNumber),
	}
	return
}
