package database

import (
	"database/sql"
	"fmt"

	"github.com/Tinyblargon/DemoOnDemand/dod/helper/programconfig"
	_ "github.com/lib/pq"
)

var db *sql.DB

type UserLinkedList struct {
	LinkedList *UserLinkedList
	UserName   string
	BindDn     string
}

func New(config programconfig.PostgreSQLConfiguration) (db *sql.DB, err error) {
	return sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Database))
}

func AddDemoOfUser(db *sql.DB, userName, demoName string, demoNumber uint) (err error) {
	_, err = db.Exec(`insert into "runningdemos"("username","demoname","demonumber","running") values($1, $2, $3, $4)`, userName, demoName, demoNumber, false)
	return
}

func DeleteDemoOfUser(db *sql.DB, userName, demoName string, demoNumber uint) (err error) {
	_, err = db.Exec(`delete from "runningdemos" where username=$1 AND demoname=$2 AND demonumber=$3`, userName, demoName, demoNumber)
	return
}

func UpdateDemoOfUser(db *sql.DB, userName, demoName string, demoNumber uint, running bool) (err error) {
	_, err = db.Exec(`update "runningdemos" set "running"=$1 where username=$2 AND demoname=$3 AND demonumber=$4`, running, userName, demoName, demoNumber)
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

func GetSpecificDemo(db *sql.DB, userName, demoName string, demoNumber uint) (demo *Demo, err error) {
	rows, err := db.Query(`SELECT "username","demoname","demonumber","running" FROM "runningdemos" WHERE username=$1 AND demoname=$2 AND demonumber=$3`, userName, demoName, demoNumber)
	if err != nil {
		return
	}
	demos, err := getDemosFromRows(rows)
	if err != nil {
		return
	}
	if demos != nil {
		if (*demos)[0] != nil {
			demo = (*demos)[0]
		}
	}
	return
}

func ListDemosOfUser(db *sql.DB, userName string) (*[]*Demo, error) {
	rows, err := db.Query(`SELECT "username","demoname","demonumber","running" FROM "runningdemos" WHERE username=$1`, userName)
	if err != nil {
		return nil, err
	}
	return getDemosFromRows(rows)
}

func ListDemosOfTemplate(db *sql.DB, template string) (*[]*Demo, error) {
	rows, err := db.Query(`SELECT "username","demoname","demonumber","running" FROM "runningdemos" WHERE demoname=$1`, template)
	if err != nil {
		return nil, err
	}
	return getDemosFromRows(rows)
}

func ListAllDemos(db *sql.DB) (*[]*Demo, error) {
	rows, err := db.Query(`SELECT "username","demoname","demonumber","running" FROM "runningdemos"`)
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

func CheckTemplateInUse(db *sql.DB, demoName string) (inUse bool, err error) {
	rows, err := db.Query(`SELECT "demoname" FROM "runningdemos" WHERE demoname=$1`, demoName)
	if err != nil {
		return
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			break
		}
		inUse = true
		break
	}
	rows.Close()
	return
}
