package db

import (
	"database/sql"
	"time"
	"types"
	"utils"

	_ "github.com/go-sql-driver/mysql" //Mysql driver
)

//MySQL - MYSQL database class
type MySQL struct {
	sql *sql.DB
}

//Init - Start pooling with mysql
func (db MySQL) Init(config *types.Config) *MySQL {
	pool, _ := sql.Open("mysql", config.MySQL.UserName+":"+config.MySQL.Password+"@tcp(["+config.MySQL.Conn+"])/"+config.MySQL.DBName+"?parseTime=true")
	if pool.Ping() != nil {
		return nil
	}
	db.sql = pool

	//Setup interval to remove expired data
	utils.Schedule(db.DeleteExpired, 1*time.Hour)

	return &db
}

//PreparedQuery - returns a prepared statement
func (db MySQL) PreparedQuery(query string) (*sql.Stmt, error) {
	stmt, err := db.sql.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

//SimpleQuery - returns the result of a query
func (db MySQL) SimpleQuery(query string) (*sql.Rows, error) {
	q, err := db.sql.Query(query)
	if err != nil {
		return nil, err
	}
	return q, nil
}

//DeleteExpired - removes all expired recoveries or devices
func (db MySQL) DeleteExpired() {
	_, _ = db.SimpleQuery("DELETE FROM recover WHERE created < (NOW() - INTERVAL 1 HOUR)")
	_, _ = db.SimpleQuery("DELETE FROM emailChange WHERE created < (NOW() - INTERVAL 1 HOUR)")
	_, _ = db.SimpleQuery("DELETE FROM devices WHERE created < (NOW() - INTERVAL 60 DAY)")
}
