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

type DemosOfUser struct {
	DemoName   string
	DemoNumber uint
}

func NumberOfDomosOfUser(db *sql.DB, userName string) (numberOfDemos uint, err error) {
	array, err := ListDemosOfUser(db, userName)
	numberOfDemos = uint(len(array))
	return
}

func ListDemosOfUser(db *sql.DB, userName string) (array []*DemosOfUser, err error) {
	rows, err := db.Query(`SELECT "username","demoname","demonumber" FROM "runningdemos" WHERE username=$1`, userName)
	if err != nil {
		return
	}
	array = make([]*DemosOfUser, 0)
	for rows.Next() {
		child := new(DemosOfUser)
		var userName string
		var demoName string
		var demoNumber uint
		err = rows.Scan(&userName, &demoName, &demoNumber)
		if err != nil {
			return
		}
		child.DemoName = demoName
		child.DemoNumber = demoNumber
		array = append(array, child)
	}
	return
}
