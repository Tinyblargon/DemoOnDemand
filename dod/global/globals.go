package global

import "strings"

var ConfigFolder string
var DataCenter string
var TemplateFodler string
var RouterFodler string
var DemoFodler string
var IngressVM string

func SetAll(dataCenter, demofolder, configFolder string) {
	ConfigFolder = configFolder

	DataCenter = dataCenter
	baseFolder := strings.Trim(demofolder, "/")
	TemplateFodler = baseFolder + "/Templates"
	RouterFodler = baseFolder + "/Router"
	DemoFodler = baseFolder + "/Demos"
	IngressVM = "routervm"
}
