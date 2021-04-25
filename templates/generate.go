package templates

import (
	"github.com/SeraphJACK/v2stat/db"
	"github.com/SeraphJACK/v2stat/util"
	"io"
	"strconv"
	"text/template"
	"time"
)

type dayRecord struct {
	Date string
	Rx   int64
	Tx   int64
}

type userRecord struct {
	User string
	Rx   int64
	Tx   int64
}

type templateData struct {
	Date                 string
	TotalTraffic         string
	TotalRx              string
	TotalTx              string
	UserCount            string
	TrafficLastMonth     string
	LastWeekRecords      []dayRecord
	UserRecords          []userRecord
	MonthlyUserRecords   []userRecord
	LastMonthUserRecords []userRecord
}

func Generate(w io.Writer) error {
	data := templateData{}
	data.Date = time.Now().Format("2006-01-02 15:04:05")

	var totalRx, totalTx, total int64
	for _, v := range db.QueryDaySum(util.ThisMonth()) {
		totalRx += v.Rx
		totalTx += v.Tx
		total += v.Rx
		total += v.Tx
	}
	data.TotalRx = util.FormatTraffic(totalRx)
	data.TotalTx = util.FormatTraffic(totalTx)
	data.TotalTraffic = util.FormatTraffic(total)

	data.UserCount = strconv.Itoa(db.UserCount())

	for i := -7; i < 0; i++ {
		t := util.Day(i)
		date := t.Format("2006-01-02")
		data.LastWeekRecords = append(data.LastWeekRecords, calcSum(db.QueryDay(t), date))
	}
	data.LastWeekRecords = append(data.LastWeekRecords,
		calcSum(db.QueryHourSum(util.Today()), util.Today().Format("2006-01-02")))

	rec := make(map[string]userRecord)
	for i := -7; i < 0; i++ {
		t := util.Day(i)
		for _, v := range db.QueryDay(t) {
			r := rec[v.User]
			r.Rx += v.Rx
			r.Tx += v.Tx
			rec[v.User] = r
		}
	}
	for _, v := range db.QueryHourSum(util.Today()) {
		r := rec[v.User]
		r.Rx += v.Rx
		r.Tx += v.Tx
		rec[v.User] = r
	}
	for k, v := range rec {
		v.User = k
		data.UserRecords = append(data.UserRecords, v)
	}

	now := time.Now()
	lastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())

	totalRx = 0
	totalTx = 0

	for _, v := range db.QueryMonth(lastMonth) {
		totalRx += v.Rx
		totalTx += v.Tx
	}

	data.TrafficLastMonth = util.FormatTraffic(totalRx + totalTx)

	for _, v := range db.QueryDaySum(util.ThisMonth()) {
		data.MonthlyUserRecords = append(data.MonthlyUserRecords, userRecord{
			User: v.User,
			Rx:   v.Rx,
			Tx:   v.Tx,
		})
	}

	for _, v := range db.QueryMonth(util.ThisMonth().AddDate(0, -1, 0)) {
		data.LastMonthUserRecords = append(data.LastMonthUserRecords, userRecord{
			User: v.User,
			Rx:   v.Rx,
			Tx:   v.Tx,
		})
	}

	tmp := template.Must(template.ParseFiles("report.gohtml"))
	return tmp.Execute(w, data)
}

func calcSum(records []db.Record, date string) dayRecord {
	var rx, tx int64
	for _, v := range records {
		rx += v.Rx
		tx += v.Tx
	}
	return dayRecord{
		Date: date,
		Rx:   rx,
		Tx:   tx,
	}
}
