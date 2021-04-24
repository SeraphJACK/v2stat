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

type templateData struct {
	Date            string
	TotalTraffic    string
	TotalRx         string
	TotalTx         string
	UserCount       string
	LastWeekRecords []dayRecord
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
