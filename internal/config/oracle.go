package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/godror/godror"
)

func InitDB() *sql.DB {

	dsn := fmt.Sprintf(
		`user="%s" password="%s" libDir="%s" timezone="%s" connectString="(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=%s)(PORT=%s))(CONNECT_DATA=(SERVICE_NAME=%s)))"`,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_LIB_DIR"),
		os.Getenv("DB_TIMEZONE"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SERVICE"),
	)

	db, err := sql.Open("godror", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db
}
