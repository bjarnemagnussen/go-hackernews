package main

import (
	"testing"
	"time"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			want: "17 Dec 2020 at 10:00",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Dec 2020 at 09:00",
		},
	} // Loop over the test cases.
	for _, tt := range tests {
		// Use the t.Run() function to run a sub-test for each test case. The
		// first parameter to this is the name of the test (which is used to
		// identify the sub-test in any log output) and the second parameter is
		// and anonymous function containing the actual test for each case.
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)
			if hd != tt.want {
				t.Errorf("want %q; got %q", tt.want, hd)
			}
		})
	}
}

func TestMoment(t *testing.T) {
	tests := []struct {
		name string
		tm1  time.Time
		tm2  time.Time
		want string
	}{
		{
			name: "Year",
			tm1:  time.Date(2019, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			want: "1 year",
		},
		{
			name: "Years",
			tm1:  time.Date(2018, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 18, 10, 0, 0, 0, time.UTC),
			want: "2 years",
		},
		{
			name: "One Month",
			tm1:  time.Date(2020, 11, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 25, 10, 0, 0, 0, time.UTC),
			want: "1 month",
		},
		{
			name: "Two Months",
			tm1:  time.Date(2020, 11, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2021, 1, 18, 10, 0, 0, 0, time.UTC),
			want: "2 months",
		},
		{
			name: "Many months",
			tm1:  time.Date(2020, 11, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2021, 1, 17, 10, 0, 0, 0, time.UTC),
			want: "2 months",
		},
		{
			name: "Day",
			tm1:  time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 18, 10, 0, 0, 0, time.UTC),
			want: "1 day",
		},
		{
			name: "Days",
			tm1:  time.Date(2020, 12, 12, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2021, 1, 11, 10, 0, 0, 0, time.UTC),
			want: "30 days",
		},
		{
			name: "Hour",
			tm1:  time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 17, 11, 0, 1, 0, time.UTC),
			want: "1 hour",
		},
		{
			name: "Hours",
			tm1:  time.Date(2020, 12, 17, 1, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 18, 0, 59, 59, 0, time.UTC),
			want: "23 hours",
		},
		{
			name: "Minute",
			tm1:  time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 17, 10, 1, 0, 0, time.UTC),
			want: "1 min",
		},
		{
			name: "Minutes",
			tm1:  time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 17, 10, 59, 59, 0, time.UTC),
			want: "59 mins",
		},
		{
			name: "Seconds",
			tm1:  time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			tm2:  time.Date(2020, 12, 17, 10, 0, 59, 0, time.UTC),
			want: "59 secs",
		},
	} // Loop over the test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := moment(tt.tm1, tt.tm2)
			if hd != tt.want {
				t.Errorf("want %q; got %q", tt.want, hd)
			}
		})
	}
}
