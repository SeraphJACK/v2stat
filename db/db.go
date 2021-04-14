package db

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
	"time"
)

var Dir = flag.String("db", "/var/lib/v2stat", "The directory to store traffic data")
var db *sql.DB

func InitDb() error {
	// make sure no suffix '/'
	*Dir = path.Clean(*Dir)

	info, err := os.Stat(*Dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", *Dir)
	}

	db, err = sql.Open("sqlite3", *Dir+"/v2stat.db?mode=ro")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `hour`\n(\n    `user` INT      NOT NULL,\n    `date` DATETIME NOT NULL,\n    `rx`   INT      NOT NULL DEFAULT 0,\n    `tx`   INT      NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS `day`\n(\n    `user` INTEGER NOT NULL,\n    `date` DATE    NOT NULL,\n    `rx`   INT     NOT NULL DEFAULT 0,\n    `tx`   INT     NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS `month`\n(\n    `user` INTEGER NOT NULL,\n    `date` DATE    NOT NULL,\n    `rx`   INT     NOT NULL DEFAULT 0,\n    `tx`   INT     NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS `user`\n(\n    `id`   INTEGER PRIMARY KEY,\n    `name` TEXT NOT NULL,\n    UNIQUE (`name`)\n);\n")
	if err != nil {
		return err
	}
	return nil
}

func Record(user string, rx int, tx int, t *time.Time) error {
	// Create user record if none
	_, err := db.Exec("INSERT INTO `user`(`name`) SELECT (?) WHERE NOT EXISTS(SELECT * FROM `user` WHERE `name`=?)", user, user)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO `hour`(`user`, `date`, `rx`, `tx`) SELECT `user`.id AS `user`, ? AS `date`, ? AS `rx`, ? AS `tx` FROM `user` WHERE `name`=?", t, rx, tx, user)
	return err
}
