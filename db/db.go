package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type DataBase struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	Gateway  string
	Conn     *sql.DB
}

type Influx struct {
	Database string
	User     string
	Password string
}

func (db *DataBase) Connect() {
	sqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true", db.User, db.Password, db.Host, db.Port, db.Database)
	conn, err := sql.Open("mysql", sqlDsn)
	if err != nil {
		panic(err)
	}
	db.Conn = conn
}

func (db *DataBase) Close() {
	db.Conn.Close()
	fmt.Println("\r=> Close Database <=")
}

func (db *DataBase) NotResultQueryExec(sql string) {
	_, err := db.Conn.Exec(sql)
	if err != nil {
		log.Println(err)
	}
}

func (db *DataBase) SelectDataInsertQuery(sql string) ([]string, map[string]interface{}) {
	var (
		tempTagName []string
	)
	mapMstDevice := make(map[string]interface{})
	rows, err := db.Conn.Query(sql)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	cols, _ := rows.Columns()
	defer rows.Close()

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))

		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			log.Println(err)
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if str, ok := (*val).([]uint8); ok {
				myString := string(str)
				num, err := strconv.ParseFloat(myString, 64)
				if err != nil || colName == "Mac" {
					m[colName] = myString
				} else {
					m[colName] = num
				}

			}
		}
		m["Item"] = int8(1)
		strTagName := fmt.Sprintf("%v.%s", m["Mac"], m["DefTable"].(string)[7:])
		tempTagName = append(tempTagName, strTagName)
		mapMstDevice[strTagName] = m

	}
	return tempTagName, mapMstDevice
}

func (db *DataBase) AlgorithmCheck(sql string) []string {
	var mac, deftable string
	var arrTagName []string
	rows, err := db.Conn.Query(sql)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&mac, &deftable); err != nil {
			log.Println(err)
		}
		TagName := fmt.Sprintf("%s.%s", mac, deftable[7:])
		arrTagName = append(arrTagName, TagName)
	}

	return arrTagName
}
