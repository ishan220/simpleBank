package main

import (
	"SimpleBank/api"
	db "SimpleBank/db/sqlc"
	"SimpleBank/db/util"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" //this pkg is important to connect to DB
)

// const driver string = "postgres"
// const datasource string = "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable"
// const serverAddress = "0.0.0.0:8080"

// in this init we are setting the envir var which will override the value being set in config file
//
//	func init() {
//		err := os.Setenv("SERVER_ADDRESS", "0.0.0.0:8081")
//		if err != nil {
//			log.Fatal("Problem in setting environment variable")
//		}
//	}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot Load the config file")
	}
	//fmt.Printf("%T", config.DBDriver)
	fmt.Println(config.DataSource)
	fmt.Println(config.ServerAddress)

	conn, err := sql.Open(config.DBDriver, config.DataSource)
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal("DataBase Connection failed")
		return
	}
	store := db.NewStore(conn) //db here is package name
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal("cannot create the server")
	}
	err1 := server.Start(config.ServerAddress)
	if err1 != nil {
		log.Fatal("Server Couldn't Start")
		return
	}
}
