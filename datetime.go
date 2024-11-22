package twigots

import (
	"encoding/json"
	"strconv"
	"time"
)

const (
	dateTimeLayout = "2006-01-02T15:04:05Z"
	dateLayout     = "2006-01-02"
	timeLayout     = "15:04:05"
)

// DateTime is a date and time.
type DateTime struct{ time.Time }

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var dateTimeString string
	err := json.Unmarshal(data, &dateTimeString)
	if err != nil {
		return err
	}

	parsedDateTime, err := time.Parse(dateTimeLayout, dateTimeString)
	if err != nil {
		return err
	}
	dt.Time = parsedDateTime
	return nil
}

// Date is a date (with no time).
type Date struct{ time.Time }

func (d *Date) UnmarshalJSON(data []byte) error {
	var dateString string
	err := json.Unmarshal(data, &dateString)
	if err != nil {
		return err
	}

	parsedDate, err := time.Parse(dateLayout, dateString)
	if err != nil {
		return err
	}
	d.Time = parsedDate
	return nil
}

// Date is a time (with no date).
type Time struct{ time.Time }

func (t *Time) UnmarshalJSON(data []byte) error {
	var timeString string
	err := json.Unmarshal(data, &timeString)
	if err != nil {
		return err
	}

	parsedTime, err := time.Parse(timeLayout, timeString)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}

// UnixTime is a time from unix time.
type UnixTime struct{ time.Time }

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	var timeString string
	err := json.Unmarshal(data, &timeString)
	if err != nil {
		return err
	}

	timeInt, err := strconv.Atoi(timeString)
	if err != nil {
		return err
	}

	t.Time = time.UnixMilli(int64(timeInt))
	return nil
}
