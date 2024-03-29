package db

import (
	"SimpleBank/db/util"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	//"testing"
	_ "github.com/lib/pq"
)

//const driver string = "postgres"
//const datasource string = "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable"

var TestQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Couldn't laod config file")
	}
	testDB, err = sql.Open(config.DBDriver, config.DataSource)
	if err != nil {
		log.Fatal("cannot connect to Db:", err)
	}
	err1 := testDB.Ping()
	if err1 != nil {
		fmt.Println("Database connection Not established")
	} else {
		fmt.Println("Database connection Successfully established")
	}
	TestQueries = New(testDB)
	if TestQueries != nil {
		fmt.Println("testQueries", TestQueries)
	}
	fmt.Println("m.Run() function is called")
	os.Exit(m.Run())

}
