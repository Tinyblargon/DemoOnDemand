package demos

import (
	"context"
	"net/http"

	demoactions "github.com/Tinyblargon/DemoOnDemand/dod/demoActions"
	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/database"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/datacenter"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/provider"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/session"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/vsphere/virtualmachine"
	"github.com/Tinyblargon/DemoOnDemand/dod/scheduler/job"
	"github.com/Tinyblargon/DemoOnDemand/dod/template"
	"github.com/gorilla/mux"
)

type StartStopRestart struct {
	Task string `json:"task"`
}

type Data struct {
	Demos *[]*Demo `json:"demos"`
}

type Demo struct {
	UserName         string                  `json:"user"`
	DemoName         string                  `json:"demo"`
	DemoNumber       uint                    `json:"number"`
	Running          bool                    `json:"active"`
	Description      string                  `json:"description,omitempty"`
	RouterConnection string                  `json:"router-ip,omitempty"`
	PortForward      []*template.PortForward `json:"portforwards,omitempty"`
}

var GetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Get(w, r)
})

var PostHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Post(w, r)
})

var IdGetHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdGet(w, r)
})

var IdDeleteHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdDelete(w, r)
})

var IdPutHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	IdPut(w, r)
})

func Get(w http.ResponseWriter, r *http.Request) {
	var demos *[]*database.Demo
	var err error
	var allDemos bool
	if r.Header.Get("role") == "root" {
		allDemos = true
		demos, err = database.ListAllDemos(global.DB)
	} else {
		demos, err = database.ListDemosOfUser(global.DB, r.Header.Get("name"))
	}
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	uniqueDemos, err := getUniqueDemo(demos)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	demoList := make([]*Demo, len(*demos))
	for i, e := range *demos {
		var userName string
		if allDemos {
			userName = e.UserName
		}
		for _, ee := range *uniqueDemos {
			if e.DemoName == ee.DemoName {
				demoList[i] = &Demo{
					UserName:    userName,
					DemoName:    e.DemoName,
					DemoNumber:  e.DemoNumber,
					Running:     e.Running,
					Description: ee.Description,
				}
			}
		}
	}
	response := api.JsonResponse{
		Data: Data{
			Demos: &demoList,
		},
	}
	response.Output(w)
}

func Post(w http.ResponseWriter, r *http.Request) {
	newDemo := job.Demo{
		Create: true,
	}
	err := api.GetBody(r, &newDemo)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}
	if newDemo.UserName == "" {
		newDemo.UserName = r.Header.Get("name")
	}
	if !api.IfRoleOrUser(r, "root", newDemo.UserName) {
		api.OutputInvalidPermission(w)
		return
	}
	demoObj := demo.Demo{
		Name: newDemo.Template,
		User: newDemo.UserName,
		ID:   newDemo.Number,
	}
	existence, err := demoactions.CheckExistence(global.DB, demoObj)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	if existence {
		api.OutputDemoAlreadyExists(w)
		return
	}
	newJob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newJob, newDemo.UserName)
}

type IdData struct {
	Demo *Demo `json:"demo"`
}

func IdGet(w http.ResponseWriter, r *http.Request) {
	demoObj, err := checkID(w, r)
	if err != nil {
		return
	}
	if !api.IfRoleOrUser(r, "root", demoObj.User) {
		api.OutputInvalidPermission(w)
		return
	}
	demo, err := database.GetSpecificDemo(global.DB, demoObj)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	existence, err := demoactions.CheckExistence(global.DB, demoObj)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	if !existence {
		api.OutputDemoDoesNotExists(w)
		return
	}
	templateConf, err := template.Get(demoObj.Name)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	networks, err := database.ListUsedNetworksOfDemo(global.DB, &demoObj)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	guestIp, err := obtainGuestIP(demoObj, networks)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	portForwards := make([]*template.PortForward, len(templateConf.PortForwards))
	for i, e := range templateConf.PortForwards {
		portForwards[i] = &template.PortForward{
			SourcePort: e.SourcePort,
			Protocol:   e.Protocol,
			Comment:    e.Comment,
		}
	}
	response := api.JsonResponse{
		Data: IdData{
			Demo: &Demo{
				UserName:         demo.UserName,
				DemoName:         demo.DemoName,
				DemoNumber:       demo.DemoNumber,
				Running:          demo.Running,
				Description:      templateConf.Description,
				RouterConnection: guestIp,
				PortForward:      portForwards,
			},
		},
	}
	response.Output(w)
}

func IdDelete(w http.ResponseWriter, r *http.Request) {
	demoObj, err := checkID(w, r)
	if err != nil {
		return
	}
	if !api.IfRoleOrUser(r, "root", demoObj.User) {
		api.OutputInvalidPermission(w)
		return
	}
	existence, err := demoactions.CheckExistence(global.DB, demoObj)
	if err != nil {
		api.OutputServerError(w, "", err)
		return
	}
	if !existence {
		api.OutputDemoDoesNotExists(w)
		return
	}
	newDemo := job.Demo{
		Template: demoObj.Name,
		UserName: demoObj.User,
		Number:   demoObj.ID,
		Destroy:  true,
	}
	newJob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newJob, demoObj.User)
}

func IdPut(w http.ResponseWriter, r *http.Request) {
	demoObj, err := checkID(w, r)
	if err != nil {
		return
	}
	if !api.IfRoleOrUser(r, "root", demoObj.User) {
		api.OutputInvalidPermission(w)
		return
	}
	SSR := StartStopRestart{}
	err = api.GetBody(r, &SSR)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}
	newDemo := job.Demo{
		Template: demoObj.Name,
		UserName: demoObj.User,
		Number:   demoObj.ID,
	}
	switch SSR.Task {
	case "start":
		newDemo.Start = true
	case "stop":
		newDemo.Stop = true
	case "restart":
		newDemo.Start = true
		newDemo.Stop = true
	default:
		api.OutputUserInputError(w, "Key task should be (Start|Stop|Restart)")
		return
	}
	newJob := job.Job{
		Demo: &newDemo,
	}
	api.NewJob(w, &newJob, demoObj.User)
}

func checkID(w http.ResponseWriter, r *http.Request) (demoObj demo.Demo, err error) {
	vars := mux.Vars(r)
	id := vars["id"]
	demoObj, err = demo.CreateObject(id)
	if err != nil {
		api.OutputInvalidID(w)
	}
	return
}

// gets the file information of every unique demo in the list
func getUniqueDemo(list *[]*database.Demo) (uniqueList *[]Demo, err error) {
	tmpList := make([]Demo, 0)
	uniqueList = &tmpList
	for _, e := range *list {
		var templateConf *template.Config
		if isDemoUnique(uniqueList, e.DemoName) {
			templateConf, err = template.Get(e.DemoName)
			if err != nil {
				return
			}
			demo := Demo{
				DemoName:    e.DemoName,
				Description: templateConf.Description,
			}
			*uniqueList = append(*uniqueList, demo)
		}
	}
	return
}

// checks if a demo with a specific name already exists in the list
func isDemoUnique(list *[]Demo, item string) bool {
	for _, e := range *list {
		if e.DemoName == item {
			return false
		}
	}
	return true
}

func obtainGuestIP(demoObj demo.Demo, networks []string) (guestIP string, err error) {
	c, err := session.New(*vsphere.GetConfig())
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.GetTimeout())
	defer cancel()
	defer c.VimClient.Logout(ctx)
	dataCenter, err := datacenter.Get(c.VimClient, datacenter.GetName())
	if err != nil {
		return
	}
	guestIP, _, err = virtualmachine.GetGuestIP(c.VimClient, demoObj.CreateDemoURl(), global.IngressVM, networks, dataCenter, nil)
	return
}
