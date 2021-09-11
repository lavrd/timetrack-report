package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
	To start tests use following command: go test -count 1 -v -cover .
*/

/*
	Add tests:
		interval 05, 10 -> 05-09/10-04
		interval 02, 05, 15, 20 -> 02-04/05-14/15-19/20-01
		interval 01, 20 -> 01-19/20-last_day_of_month
		interval 01, last_day_of_month -> 01-last_day_of_month
		interval 10, last_day_of_month -> 10-last_day_of_month-1/last_day_of_month-09
*/

func Test_getDates(t *testing.T) {
	type args struct {
		date time.Time
	}
	cases := []struct {
		name      string
		args      args
		startDate string
		endDate   string
	}{
		{
			name: "10, 25 -> 10-24/25-09 : 2021-02-11",
			args: args{
				date: time.Date(2021, 02, 11, 0, 0, 0, 0, time.Local),
			},
			startDate: "2021-02-10",
			endDate:   "2021-02-24",
		},
		{
			name: "10, 25 -> 10-24/25-09 : 2021-02-05",
			args: args{
				date: time.Date(2021, 02, 05, 0, 0, 0, 0, time.Local),
			},
			startDate: "2021-01-25",
			endDate:   "2021-02-09",
		},
		{
			name: "10, 25 -> 10-24/25-09 : 2021-02-26",
			args: args{
				date: time.Date(2021, 02, 26, 0, 0, 0, 0, time.Local),
			},
			startDate: "2021-02-25",
			endDate:   "2021-03-09",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			startDate, endDate := getDates(c.args.date)
			assert.Equal(t, c.startDate, startDate)
			assert.Equal(t, c.endDate, endDate)
		})
	}
}
