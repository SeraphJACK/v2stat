package main

import (
	"flag"
	"fmt"
	"github.com/SeraphJACK/v2stat/config"
	"github.com/SeraphJACK/v2stat/db"
	"os"
	"strconv"
	"time"
)

var units = map[int]string{
	0: "B",
	1: "KiB",
	2: "MiB",
	3: "GiB",
	4: "TiB",
}

func formatTraffic(traffic int64) string {
	unit := 0
	num := float64(traffic)
	for unit <= 3 && num >= 1000 {
		unit++
		num /= 1024
	}
	return strconv.FormatFloat(num, 'f', 2, 64) + " " + units[unit]
}

func printRecords(title string, records []db.Record) {
	fmt.Printf("Traffics for %s :\n", title)
	if len(records) == 0 {
		fmt.Print("    not available\n\n")
		return
	}
	fmt.Printf("    %-15v / %-10v / %-10v\n", "user", "rx", "tx")
	sumRx := int64(0)
	sumTx := int64(0)
	for _, v := range records {
		sumRx += v.Rx
		sumTx += v.Tx
		fmt.Printf("    %-15v / %-10v / %-10v\n", v.User, formatTraffic(v.Rx), formatTraffic(v.Tx))
	}
	fmt.Print("\n")

	fmt.Printf("    Sum rx: %s, tx: %s, total: %s\n", formatTraffic(sumRx), formatTraffic(sumTx), formatTraffic(sumRx+sumTx))

	fmt.Print("\n")
}

func main() {
	prevDays := flag.Int("days", 5, "Amount of days before today to show")
	prevMonths := flag.Int("months", 1, "Amount of months before this month to show")
	showToday := flag.Bool("today", true, "Show today")
	showThisMonth := flag.Bool("thisMonth", false, "Show this month")

	flag.Parse()
	err := config.LoadConf()
	if err != nil {
		fmt.Printf("Failed to parse config: %v\n", err)
		os.Exit(1)
	}
	err = db.InitDb(true)
	if err != nil {
		fmt.Printf("Failed to connect to db: %v\n", err)
		os.Exit(1)
	}

	now := time.Now()

	for i := *prevMonths; i > 0; i-- {
		t := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, now.Location())
		title := strconv.Itoa(int(t.Month())) + "/" + strconv.Itoa(t.Year())
		printRecords(title, db.QueryMonth(t))
	}

	if *showThisMonth {
		title := "this month"
		printRecords(title, db.QueryDaySum(now))
	}

	for i := *prevDays; i > 0; i-- {
		t := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
		title := strconv.Itoa(t.Day()) + "/" + strconv.Itoa(int(t.Month())) + "/" + strconv.Itoa(t.Year())
		printRecords(title, db.QueryDay(t))
	}

	if *showToday {
		title := "today"
		printRecords(title, db.QueryHourSum(now))
	}
}
