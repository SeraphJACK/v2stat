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
	t := time.Now()

	if config.Config.Debug {
		log.Printf("Starting to do record, current time: %s\n", t.String())
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
		err = db.Record(k, v.rx, v.tx, t)
		if err != nil {
			log.Printf("Failed to store record into db: %v\n", err)
		}
	}
}

func main() {
	err := config.LoadConf()
	if err != nil {
		panic(err)
	}

	err = db.InitDb()
	if err != nil {
		panic(err)
	}

	_, _ = util.QueryStats(config.Config.ServerAddr, "", true)
	c := cron.New()

	_, err = c.AddFunc("0 * * * *", DoRecord)
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc("0 0 * * *", func() {
		err := db.SumDay(time.Now())
		if err != nil {
			log.Printf("Failed to sum day: %v\n", err)
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = c.AddFunc("0 0 1 * *", func() {
		err := db.SumMonth(time.Now())
		if err != nil {
			log.Printf("Failed to sum month: %v\n", err)
		}
	})
	if err != nil {
		panic(err)
	}

	c.Start()
}
