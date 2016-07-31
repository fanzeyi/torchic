package mysql

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

var db *sql.DB

func mysqlDSN(address, username, password, dbname string) string {
	config := mysql.Config{
		User:   username,
		Passwd: password,
		Net:    "tcp",
		Addr:   address,
		DBName: dbname,
	}

	return config.FormatDSN()
}

func Init(address, username, password, dbname string) {
	var err error

	db, err = sql.Open("mysql", mysqlDSN(address, username, password, dbname))

	if err != nil {
		glog.Errorf("Error when connecting to MySQL server: %s", err)
		return
	}

	glog.Infof("Connected to MySQL")
}

func GetConn() *sql.DB {
	return db
}
