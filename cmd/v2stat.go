package main

import (
	"flag"
	"fmt"
	"github.com/SeraphJACK/v2stat/config"
	"github.com/SeraphJACK/v2stat/db"
	"github.com/SeraphJACK/v2stat/util"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func printRecords(title string, records []db.Record) {
	fmt.Printf("Traffics for %s :\n", title)
	if len(records) == 0 {
		fmt.Print("    not available\n\n")
		return
	}
	fmt.Printf("    +-----------------+------------+-------------+\n")
	fmt.Printf("    | %-15v | %-10v | %-10v |\n", "user", "rx", "tx")
	fmt.Printf("    +-----------------+------------+-------------+\n")
	sumRx := int64(0)
	sumTx := int64(0)
	for _, v := range records {
		sumRx += v.Rx
		sumTx += v.Tx
		fmt.Printf(
			"    | %-15v | %-10v | %-10v |\n",
			v.User,
			util.FormatTraffic(v.Rx),
			util.FormatTraffic(v.Tx),
		)
	}
	fmt.Printf("    +-----------------+------------+-------------+\n")
	fmt.Print("\n")

	fmt.Printf(
		"    Sum rx: %s, tx: %s, total: %s\n",
		util.FormatTraffic(sumRx),
		util.FormatTraffic(sumTx),
		util.FormatTraffic(sumRx+sumTx),
	)

	fmt.Print("\n")
}

func cliStats(prevDays, prevMonths int, showToday, showThisMonth bool) {
	now := time.Now()

	for i := prevMonths; i > 0; i-- {
		t := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, now.Location())
		title := strconv.Itoa(int(t.Month())) + "/" + strconv.Itoa(t.Year())
		printRecords(title, db.QueryMonth(t))
	}

	if showThisMonth {
		title := "this month"
		printRecords(title, db.QueryDaySum(now))
	}

	for i := prevDays; i > 0; i-- {
		t := time.Date(now.Year(), now.Month(), now.Day()-i, 0, 0, 0, 0, now.Location())
		title := strconv.Itoa(int(t.Month())) + ", " + strconv.Itoa(t.Day()) + ", " + strconv.Itoa(t.Year())
		printRecords(title, db.QueryDay(t))
	}

	if showToday {
		title := "today"
		printRecords(title, db.QueryHourSum(now))
	}
}

func main() {
	prevDays := flag.Int("days", 5, "Amount of days before today to show")
	prevMonths := flag.Int("months", 1, "Amount of months before this month to show")
	showToday := flag.Bool("today", true, "Show today")
	showThisMonth := flag.Bool("thisMonth", false, "Show this month")
	htmlOutputPath := flag.String("output", "report.html", "Path to write html report")

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

	skip := false
	for _, cmd := range os.Args[1:] {
		if skip {
			skip = false
			continue
		}
		if strings.HasPrefix(cmd, "-") {
			skip = !strings.Contains(cmd, "=")
			continue
		}
		if cmd == "genreport" {
			log.Printf("Writing report to %s...\n", *htmlOutputPath)
			file, err := os.Create(*htmlOutputPath)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to write report: %v\n", err)
			}
			err = Generate(file)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to generate report: %v\n", err)
			}
			err = file.Close()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Failed to close file: %v\n", err)
			}
			os.Exit(0)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Unkonwn command: %s\n", cmd)
			os.Exit(1)
		}
	}

	cliStats(*prevDays, *prevMonths, *showToday, *showThisMonth)
	os.Exit(0)
}
