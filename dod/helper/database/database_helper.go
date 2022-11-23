package database

import (
	"database/sql"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/demo"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/name"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	_ "github.com/lib/pq"
)

type UserLinkedList struct {
	LinkedList *UserLinkedList
	UserName   string
	BindDn     string
}

type Vlan struct {
	ID     uint
	Demo   string
	Prefix string
}

func (v Vlan) GetNetwork() string {
	return name.Network(v.Prefix, v.ID)
}

func New(config programconfig.PostgreSQLConfiguration) (db *sql.DB, err error) {
	return sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Database))
}

func AddDemoOfUser(db *sql.DB, demoObj *demo.Demo) (demoID uint, err error) {
	err = db.QueryRow(`INSERT INTO "runningdemos"("username","demoname","running") VALUES ($1, $2, $3) RETURNING id`, demoObj.User, demoObj.Name, false).Scan(&demoID)
	return
}

func DeleteDemoOfUser(db *sql.DB, demoObj *demo.Demo) (err error) {
	_, err = db.Exec(`delete from "runningdemos" where username=$1 AND demoname=$2 AND id=$3`, demoObj.User, demoObj.Name, demoObj.ID)
	return
}

func UpdateDemoOfUser(db *sql.DB, demoObj *demo.Demo, running bool) (err error) {
	_, err = db.Exec(`update "runningdemos" set "running"=$1 where username=$2 AND demoname=$3 AND id=$4`, running, demoObj.User, demoObj.Name, demoObj.ID)
	return
}

func NumberOfDomosOfUser(db *sql.DB, userName string) (numberOfDemos uint, err error) {
	array, err := ListDemosOfUser(db, userName)
	numberOfDemos = uint(len(*array))
	return
}

type Demo struct {
	UserName   string `json:"user"`
	DemoName   string `json:"demo"`
	DemoNumber uint   `json:"demonumber"`
	Running    bool   `json:"active"`
}

func GetSpecificDemo(db *sql.DB, demoObj demo.Demo) (demo *Demo, err error) {
	rows, err := db.Query(`SELECT "username","demoname","id","running" FROM "runningdemos" WHERE username=$1 AND demoname=$2 AND id=$3`, demoObj.User, demoObj.Name, demoObj.ID)
	if err != nil {
		return
	}
	demos, err := getDemosFromRows(rows)
	if err != nil {
		return
	}
	if demos != nil {
		if len(*demos) > 0 {
			demo = (*demos)[0]
		}
	}
	return
}

func ListDemosOfUser(db *sql.DB, userName string) (*[]*Demo, error) {
	rows, err := db.Query(`SELECT "username","demoname","id","running" FROM "runningdemos" WHERE username=$1`, userName)
	if err != nil {
		return nil, err
	}
	return getDemosFromRows(rows)
}

func ListDemosOfTemplate(db *sql.DB, template string) (*[]*Demo, error) {
	rows, err := db.Query(`SELECT "username","demoname","id","running" FROM "runningdemos" WHERE demoname=$1`, template)
	if err != nil {
		return nil, err
	}
	return getDemosFromRows(rows)
}

func ListAllDemos(db *sql.DB) (*[]*Demo, error) {
	rows, err := db.Query(`SELECT "username","demoname","id","running" FROM "runningdemos"`)
	if err != nil {
		return nil, err
	}
	return getDemosFromRows(rows)
}

func getDemosFromRows(rows *sql.Rows) (demos *[]*Demo, err error) {
	demoList := make([]*Demo, 0)
	for rows.Next() {
		demo := new(Demo)
		err := rows.Scan(&demo.UserName, &demo.DemoName, &demo.DemoNumber, &demo.Running)
		if err != nil {
			break
		}
		demoList = append(demoList, demo)
	}
	demos = &demoList
	rows.Close()
	return
}

func ListUsedVlans(db *sql.DB) (*[]Vlan, error) {
	rows, err := db.Query(`SELECT "prefix","id","demo" FROM "vlans"`)
	if err != nil {
		return nil, err
	}
	return getVlansFromRows(rows)
}

func ListUsedNetworksOfDemo(db *sql.DB, demoObj *demo.Demo) ([]string, error) {
	vlans, err := ListUsedVlansOfDemo(db, demoObj)
	if err != nil {
		return nil, err
	}
	networks := make([]string, len(*vlans))
	for i, e := range *vlans {
		networks[i] = e.GetNetwork()
	}
	return networks, nil
}

func ListUsedVlansOfDemo(db *sql.DB, demoObj *demo.Demo) (*[]Vlan, error) {
	rows, err := db.Query(`SELECT "prefix","id","demo" FROM "vlans" WHERE demo=$1`, demoObj.CreateID())
	if err != nil {
		return nil, err
	}
	return getVlansFromRows(rows)
}

func SetVlanInUse(db *sql.DB, id uint, prefix string, demoObj *demo.Demo) (err error) {
	_, err = db.Exec(`INSERT INTO "vlans"("prefix","id","demo") VALUES($1, $2, $3)`, prefix, id, demoObj.CreateID())
	return
}

func DeleteVlanInUse(db *sql.DB, demoObj *demo.Demo) (err error) {
	_, err = db.Exec(`DELETE FROM "vlans" WHERE demo=$1`, demoObj.CreateID())
	return
}

func getVlansFromRows(rows *sql.Rows) (vlans *[]Vlan, err error) {
	vlanList := make([]Vlan, 0)
	for rows.Next() {
		vlan := Vlan{}
		err := rows.Scan(&vlan.Prefix, &vlan.ID, &vlan.Demo)
		if err != nil {
			break
		}
		vlanList = append(vlanList, vlan)
	}
	vlans = &vlanList
	rows.Close()
	return
}

func CheckTemplateInUse(db *sql.DB, demoName string) (inUse bool, err error) {
	rows, err := db.Query(`SELECT "demoname" FROM "runningdemos" WHERE demoname=$1`, demoName)
	if err != nil {
		return
	}
	inUse = rows.Next()
	rows.Close()
	return
}

func CountTemplateInUse(db *sql.DB, demoName string) (demos uint, err error) {
	rows, err := db.Query(`SELECT COUNT(*) FROM "runningdemos" WHERE demoname=$1`, demoName)
	if err != nil {
		return
	}
	for rows.Next() {
		err = rows.Scan(&demos)
		if err != nil {
			return
		}
	}
	return
}
