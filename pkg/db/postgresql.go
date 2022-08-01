package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"k8s.io/klog"
)

type DBMS struct {
	user     string
	password string
	host     string
	port     int
	db       string
}

//DB접속정보를 가지고 DB객체를 생성합니다.
func NewDBMS(user string, password string, host string, port int, db string) DBMS {
	return DBMS{user, password, host, port, db}
}

//쿼리하고 결과를 반환합니다.
func (dbms DBMS) SQLExecQuery(query string) ([]map[string]interface{}, error) {
	//DB의 접속정보를 저장하고 db를 인스턴스화 한다.
	// db, err := sql.Open("postgres", dbms.user+":"+dbms.password+"@tcp("+dbms.host+":"+dbms.port+")/"+dbms.db)
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbms.host, dbms.port, dbms.user, dbms.password, dbms.db)
	db, err := sql.Open("postgres", psqlconn)
	//함수 종료시 db를 Close한다.
	defer db.Close()

	if err != nil {
		return nil, err
	}

	//db를 통해 sql문을 실행 시킨다.
	rows, err := db.Query(query)
	// 함수가 종료되면 rows도 Close한다.
	defer rows.Close()

	//컬럼을 받아온다.
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	data := make([]interface{}, len(cols))

	for i, _ := range data {
		var d []byte
		data[i] = &d
	}

	results := make([]map[string]interface{}, 0)

	for rows.Next() {
		err := rows.Scan(data...)
		if err != nil {
			return nil, err
		}
		result := make(map[string]interface{})
		for i, item := range data {
			result[cols[i]] = string(*(item.(*[]byte)))
		}
		results = append(results, result)
	}

	return results, nil
}

func (dbms DBMS) InputData(i interface{}) error {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbms.host, dbms.port, dbms.user, dbms.password, dbms.db)
	db, err := gorm.Open(sqlite.Open(psqlconn), &gorm.Config{})
	if err != nil {
		klog.Error(err, "opening DB is failed")
		return err
	}

	if err = db.AutoMigrate(i); err != nil {
		klog.Error(err, "automigrating is failed")
		return err
	}
	db.Create(i)

	return nil
}

// func (dbms DBMS) ReadData(pk string, getter interface{}) error {
// 	db, err := gorm.Open(sqlite.Open(dbms.user+":"+dbms.password+"@tcp("+dbms.host+":"+dbms.port+")/"+dbms.db), &gorm.Config{})
// 	if err != nil {
// 		klog.Error(err, "opening DB is failed")
// 		return err
// 	}

// 	db.First(getter, pk)

// 	return nil
// }
