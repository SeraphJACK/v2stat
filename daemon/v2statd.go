package main

import (
	"github.com/SeraphJACK/v2stat/config"
	"github.com/SeraphJACK/v2stat/db"
	"github.com/SeraphJACK/v2stat/util"
	"github.com/robfig/cron/v3"
	"log"
	"strings"
	"time"
)

type record struct {
	rx int64
	tx int64
}

func DoRecord() {
	now := time.Now()

	if config.Config.Debug {
		log.Printf("Starting to do record, current time: %s\n", now.String())
	}

	res, err := util.QueryStats(config.Config.ServerAddr, "user", true)
	if err != nil {
		log.Printf("Failed to query stats: %v\n", err)
		return
	}

	records := make(map[string]record)

	for _, v := range res.Stat {
		// Ignore non-user stat entries
		if !strings.HasPrefix(v.Name, "user") {
			if config.Config.Debug {
				log.Printf("Ignored non-user stat entry: %s\n", v.Name)
			}
			continue
		}
		if strings.HasSuffix(v.Name, "downlink") {
			name := util.ExtractUser(v.Name)
			rec := records[name]
			rec.rx = v.Value
			records[name] = rec
		} else if strings.HasSuffix(v.Name, "uplink") {
			name := util.ExtractUser(v.Name)
			rec := records[name]
			rec.tx = v.Value
			records[name] = rec
		} else {
			log.Printf("Unrecognized stat entry: %s\n", v.Name)
			continue
		}
	}

	for k, v := range records {
		if config.Config.Debug {
			log.Printf("Recording %s: %d %d\n", k, v.rx, v.tx)
		}
		err = db.DoRecord(k, v.rx, v.tx, now)
		if err != nil {
			log.Printf("Failed to store record into db: %v\n", err)
		}
	}

	// Clean hour records
	before := time.Date(now.Year(), now.Month(), now.Day()-config.Config.DaysToKeep, 0, 0, 0, 0, now.Location())
	db.CleanHoursRecord(before)
}

func main() {
	err := config.LoadConf()
	if err != nil {
		panic(err)
	}

	err = db.InitDb(false)
	if err != nil {
		panic(err)
	}

	_, err = util.QueryStats(config.Config.ServerAddr, "", true)
	if err != nil {
		panic(err)
	}
	c := cron.New()

	_, err = c.AddFunc("0 * * * *", DoRecord)
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc("0 0 * * *", func() {
		now := time.Now()
		err := db.SumDay(now)
		if err != nil {
			log.Printf("Failed to sum day: %v\n", err)
		}

		// Clean day records
		before := time.Date(now.Year(), now.Month()-time.Month(config.Config.MonthsToKeep), 1, 0, 0, 0, 0, now.Location())
		db.CleanDayRecords(before)
	})
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc("0 0 1 * *", func() {
		now := time.Now()
		err := db.SumMonth(now)
		if err != nil {
			log.Printf("Failed to sum month: %v\n", err)
		}

		// Clean month records
		before := time.Date(now.Year()-config.Config.YearsToKeep, 1, 1, 0, 0, 0, 0, now.Location())
		db.CleanMonthRecords(before)
	})
	if err != nil {
		panic(err)
	}

	c.Start()

	for true {
		time.Sleep(time.Hour)
	}
}
