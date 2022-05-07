package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	mssql "github.com/denisenkom/go-mssqldb"
	"io"
	"log"
	"net/http"
)

var (
	dbHost    = flag.String("host", "localhost", "the database server")
	dbPort    = flag.Int("port", 1433, "the database port")
	dbName    = flag.String("db", "", "the database name")
	dbUser    = flag.String("user", "", "the database user")
	dbPass    = flag.String("password", "", "the database password")
	copaUrl   = flag.String("url", "https://api.copastc.io", "http[s]://host[:port]")
	copaToken = flag.String("token", "", "API token")
)

type User struct {
	Name    string `json:"first_name"`
	Surname string `json:"last_name"`
}

func getDb() *sql.DB {
	connString := fmt.Sprintf(
		"server=%s;user id=%s;password=%s;port=%d;database=%s", *dbHost, *dbUser, *dbPass, *dbPort, *dbName)
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatalln("Database connection failed:", err.Error())
	}
	return conn
}

func requestData() *http.Request {
	req, err := http.NewRequest("GET", *copaUrl+"/globalspeed/users", nil)
	if err != nil {
		log.Fatalln("Retrieving data failed:", err.Error())
	}
	req.Header.Add("Authorization", *copaToken)
	return req
}

func main() {
	flag.Parse()

	client := &http.Client{}
	resp, err := client.Do(requestData())
	if err != nil {
		log.Fatalln("API connection failed:", err.Error())
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Parse data failed:", err.Error())
	}

	var users []User
	json.Unmarshal(b, &users)

	importStr := mssql.CopyIn(
		"dbo.tblUser",
		mssql.BulkOptions{},
		"PIN", "Name", "Surname", "id_tblAdminSettings", "Role", "PasswordHashed", "HasPicture",
		"HasProfilePicture", "IsActiv", "LanguageID", "IsTrainer")
	db := getDb()

	txn, err := db.Begin()
	if err != nil {
		log.Fatalln(err.Error())
	}

	stmt, _ := txn.Prepare(importStr)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, user := range users {
		_, err = stmt.Exec("4654", user.Name, user.Surname, 50, 20, 0, 0, 0, 1, 3, 0)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	result, err := stmt.Exec()
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = stmt.Close()
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = txn.Commit()
	if err != nil {
		log.Fatalln(err.Error())
	}

	rowCount, _ := result.RowsAffected()
	log.Printf("%d rows imported\n", rowCount)
}
