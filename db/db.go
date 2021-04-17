package db

import (
	"database/sql"
	"fmt"
	"github.com/SeraphJACK/v2stat/config"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path"
	"time"
)

var db *sql.DB

type Record struct {
	User string `json:"user"`
	Rx   int64  `json:"rx"`
	Tx   int64  `json:"tx"`
}

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

// DoRecord should be invoked at the begin of the first minute of an hour in order to account the traffic
// during last hour.
func DoRecord(user string, rx int64, tx int64, t time.Time) error {
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
	if config.Config.Debug {
		log.Printf("Doing sum day, now=%s, begin=%s, end=%s\n", t.String(), begin.String(), end.String())
	}
	_, err := db.Exec(
		"INSERT INTO `day`(user, date, rx, tx)\nSELECT `user` as `user`, ? as `date`, sum(`hour`.`rx`) AS `rx`, sum(`hour`.tx) AS tx\nFROM `hour`\nWHERE `hour`.`date` >= ?\n  AND `hour`.`date` < ?\nGROUP BY `user`\n",
		begin,
		begin,
		end,
	)
	return err
}

// SumMonth Should be invoked at the first day of a month,
// the time value should be the first day after the month to sum.
func SumMonth(t time.Time) error {
	begin := time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, t.Location())
	end := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	if config.Config.Debug {
		log.Printf("Doing sum month, now=%s, begin=%s, end=%s\n", t.String(), begin.String(), end.String())
	}
	_, err := db.Exec(
		"INSERT INTO `month`(user, date, rx, tx)\nSELECT `user` as `user`, ? as `date`, sum(`day`.`rx`) AS `rx`, sum(`day`.tx) AS tx\nFROM `day`\nWHERE `day`.`date` >= ?\n  AND `day`.`date` < ?\nGROUP BY `user`\n",
		begin,
		begin,
		end,
	)
	return err
}

func QueryHourSum(day time.Time) []Record {
	begin := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := time.Date(day.Year(), day.Month(), day.Day(), 24, 0, 0, 0, day.Location())
	ret := make([]Record, 0)
	rows, err := db.Query(
		"SELECT u.name, SUM(d.rx), SUM(d.tx)\nFROM user u\n         INNER JOIN hour d ON d.user = u.id\nWHERE d.date >= ?\n  AND d.date < ?\nGROUP BY u.name\n",
		begin,
		end,
	)
	if err != nil {
		log.Printf("Failed to query day: %v\n", err)
		return ret
	}
	for rows.Next() {
		rec := Record{}
		err := rows.Scan(&rec.User, &rec.Rx, &rec.Tx)
		if err != nil {
			log.Printf("Failed to scan day query row: %v\n", err)
			continue
		}
		ret = append(ret, rec)
	}
	return ret
}

func QueryDaySum(month time.Time) []Record {
	begin := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	end := time.Date(month.Year(), month.Month()+1, 1, 24, 0, 0, 0, month.Location())
	ret := make([]Record, 0)
	rows, err := db.Query(
		"SELECT u.name, d.rx, d.tx\nFROM user u\n         INNER JOIN day d ON d.user = u.id\nWHERE d.date >= ?\n  AND d.date < ?\n",
		begin,
		end,
	)
	if err != nil {
		log.Printf("Failed to query day: %v\n", err)
		return ret
	}
	for rows.Next() {
		rec := Record{}
		err := rows.Scan(&rec.User, &rec.Rx, &rec.Tx)
		if err != nil {
			log.Printf("Failed to scan day query row: %v\n", err)
			continue
		}
		ret = append(ret, rec)
	}
	return ret
}

func QueryDay(t time.Time) []Record {
	begin := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	ret := make([]Record, 0)
	rows, err := db.Query(
		"SELECT u.name, d.rx, d.tx\nFROM user u\n         INNER JOIN day d ON d.user = u.id\nWHERE d.date = ?\n",
		begin,
	)
	if err != nil {
		log.Printf("Failed to query day: %v\n", err)
		return ret
	}
	for rows.Next() {
		rec := Record{}
		err := rows.Scan(&rec.User, &rec.Rx, &rec.Tx)
		if err != nil {
			log.Printf("Failed to scan day query row: %v\n", err)
			continue
		}
		ret = append(ret, rec)
	}
	return ret
}

func QueryMonth(t time.Time) []Record {
	begin := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	ret := make([]Record, 0)
	rows, err := db.Query(
		"SELECT u.name, d.rx, d.tx\nFROM user u\n         INNER JOIN month d ON d.user = u.id\nWHERE d.date = ?\n",
		begin,
	)
	if err != nil {
		log.Printf("Failed to query day: %v\n", err)
		return ret
	}
	for rows.Next() {
		rec := Record{}
		err := rows.Scan(&rec.User, &rec.Rx, &rec.Tx)
		if err != nil {
			log.Printf("Failed to scan day query row: %v\n", err)
			continue
		}
		ret = append(ret, rec)
	}
	return ret
}

func CleanHoursRecord(before time.Time) {
	_, err := db.Exec("DELETE FROM hour WHERE date < ?", before)
	if err != nil {
		log.Printf("Failed to clean hour records: %v\n", err)
	}
}

func CleanDayRecords(before time.Time) {
	_, err := db.Exec("DELETE FROM day WHERE date < ?", before)
	if err != nil {
		log.Printf("Failed to clean day records: %v\n", err)
	}
}

func CleanMonthRecords(before time.Time) {
	_, err := db.Exec("DELETE FROM month WHERE date < ?", before)
	if err != nil {
		log.Printf("Failed to clean month records: %v\n", err)
	}
}

func Close() error {
	return db.Close()
}
