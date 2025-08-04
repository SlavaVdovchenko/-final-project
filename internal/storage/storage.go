package storage

import (
	"database/sql"
	"os"

	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

func GetConnect() (*sql.DB, error) {
	dbFile := "scheduler.db"
	_, err := os.Stat(dbFile)

	install := true
	if err != nil {
		install = true
	}

	conn, err := sql.Open("sqlite", "file:"+dbFile)
	if err != nil {
		return nil, errors.Wrap(err, "db connect")
	}

	if install {
		schema := `
		CREATE TABLE IF NOT EXISTS scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT "",
			title VARCHAR(255) NOT NULL,
    		comment TEXT,
    		repeat VARCHAR(128)
		);
		CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
		`

		if _, err := conn.Exec(schema); err != nil {
			conn.Close()
			return nil, errors.Wrap(err, "create table down")
		}
	}

	return conn, nil
}
