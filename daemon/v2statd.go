package main

import (
	"github.com/SeraphJACK/v2stat/config"
	"github.com/SeraphJACK/v2stat/db"
	"github.com/SeraphJACK/v2stat/util"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	_, err = c.AddFunc("1 0 * * *", func() {
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

	_, err = c.AddFunc("2 0 1 * *", func() {
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	for true {
		sig := <-sigChan
		if sig == syscall.SIGHUP {
			log.Printf("Reloading configuration...\n")
			err := config.LoadConf()
			if err != nil {
				log.Printf("Failed to reload configuration: %v\n", err)
			}
		} else if sig == syscall.SIGTERM || sig == syscall.SIGINT || sig == syscall.SIGQUIT {
			// Wait all jobs to terminate
			<-c.Stop().Done()
			err := db.Close()
			if err != nil {
				log.Printf("Failed to close db connection: %v\n", err)
			}
			os.Exit(0)
		} else if sig == syscall.SIGKILL {
			os.Exit(1)
		}
	}
}
