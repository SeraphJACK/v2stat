package main

import (
	"github.com/SeraphJACK/v2stat/db"
	"github.com/SeraphJACK/v2stat/util"
	"github.com/robfig/cron/v3"
	"log"
	"strings"
	"time"
)

const addr = "127.0.0.1:10085"

type record struct {
	rx int64
	tx int64
}

func main() {
	_, _ = util.QueryStats(addr, "", true)
	c := cron.New()

	_, err := c.AddFunc("0 * * * *", func() {
		t := time.Now()

		res, err := util.QueryStats(addr, "user", true)
		if err != nil {
			log.Printf("Failed to query stats: %v\n", err)
			return
		}

		records := make(map[string]record)

		for _, v := range res.Stat {
			if strings.Contains(v.Name, "rx") {
				name := util.ExtractUser(v.Name)
				rec := records[name]
				rec.rx = v.Value
				records[name] = rec
			} else if strings.Contains(v.Name, "tx") {
				name := util.ExtractUser(v.Name)
				rec := records[name]
				rec.tx = v.Value
				records[name] = rec
			} else {
				log.Printf("Unrecognized stat name: %s\n", v.Name)
			}
		}

		for k, v := range records {
			err = db.Record(k, v.rx, v.tx, t)
			if err != nil {
				log.Printf("Failed to store record into db: %v\n", err)
			}
		}
	})
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
