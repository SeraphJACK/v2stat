package db

import (
	"database/sql"
	"fmt"
	"github.com/SeraphJACK/v2stat/config"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
	"time"
)

var db *sql.DB

func InitDb(ro bool) error {
	// make sure no suffix '/'
	Dir := path.Clean(config.Config.DbDir)

	info, err := os.Stat(Dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", Dir)
	}

	if ro {
		db, err = sql.Open("sqlite3", Dir+"/v2stat.db?mode=ro")
	} else {
		db, err = sql.Open("sqlite3", Dir+"/v2stat.db?mode=rw")
	}
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `hour`\n(\n    `user` INT      NOT NULL,\n    `date` DATETIME NOT NULL,\n    `rx`   INT      NOT NULL DEFAULT 0,\n    `tx`   INT      NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS `day`\n(\n    `user` INTEGER  NOT NULL,\n    `date` DATETIME NOT NULL,\n    `rx`   INT      NOT NULL DEFAULT 0,\n    `tx`   INT      NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS `month`\n(\n    `user` INTEGER  NOT NULL,\n    `date` DATETIME NOT NULL,\n    `rx`   INT      NOT NULL DEFAULT 0,\n    `tx`   INT      NOT NULL DEFAULT 0\n);\nCREATE TABLE IF NOT EXISTS `user`\n(\n    `id`   INTEGER PRIMARY KEY,\n    `name` TEXT NOT NULL,\n    UNIQUE (`name`)\n);\n")
	if err != nil {
		return err
	}
	return nil
}

// Record should be invoked at the begin of the first minute of an hour in order to account the traffic
// during last hour.
func Record(user string, rx int64, tx int64, t time.Time) error {
	round := time.Date(t.Year(), t.Month(), t.Day(), t.Hour()-1, 0, 0, 0, t.Location())
	// Create user record if none
	_, err := db.Exec("INSERT INTO `user`(`name`) SELECT (?) WHERE NOT EXISTS(SELECT * FROM `user` WHERE `name`=?)", user, user)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO `hour`(`user`, `date`, `rx`, `tx`)\nSELECT `user`.id AS `user`, ? AS `date`, ? AS `rx`, ? AS `tx`\nFROM `user`\nWHERE `name` = ?\n",
		round,
		rx,
		tx,
		user,
	)
	return err
}

// SumDay should be invoked at the first hour of the day,
// the time value should be the first hour after the day to sum.
func SumDay(t time.Time) error {
	begin := time.Date(t.Year(), t.Month(), t.Day()-1, 0, 0, 0, 0, t.Location())
	end := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	_, err := db.Exec(
		"INSERT INTO `day`(user, date, rx, tx)\nSELECT `user` as `user`, ? as `date`, sum(`hour`.`rx`) AS `rx`, sum(`hour`.tx) AS tx\nFROM `hour`\nWHERE `hour`.`date` >= ?\n  AND `hour`.`date` <= ?\nGROUP BY `user`\n",
		begin,
		end,
		begin,
	)
	return err
}

// SumMonth Should be invoked at the first day of a month,
// the time value should be the first day after the month to sum.
func SumMonth(t time.Time) error {
	begin := time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, t.Location())
	end := time.Date(t.Year(), t.Month(), 0, 0, 0, 0, 0, t.Location())
	_, err := db.Exec(
		"INSERT INTO `day`(user, date, rx, tx)\nSELECT `user` as `user`, ? as `date`, sum(`day`.`rx`) AS `rx`, sum(`day`.tx) AS tx\nFROM `day`\nWHERE `day`.`date` >= ?\n  AND `day`.`date` <= ?\nGROUP BY `user`\n",
		begin,
		end,
		begin,
	)
	return err
}
