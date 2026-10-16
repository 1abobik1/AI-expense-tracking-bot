package models

import "time"

const dateLayout = "2006-01-02"

// Date — дата без времени и пояса.
type Date struct {
	t time.Time // всегда полночь UTC
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func DateOf(t time.Time) Date {
	return NewDate(t.Date())
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, err
	}

	return Date{t}, nil
}

func (d Date) String() string {
	if d.IsZero() {
		return ""
	}

	return d.t.Format(dateLayout)
}

func (d Date) IsZero() bool           { return d.t.IsZero() }
func (d Date) AddDays(n int) Date     { return Date{d.t.AddDate(0, 0, n)} }
func (d Date) AddMonths(n int) Date   { return Date{d.t.AddDate(0, n, 0)} }
func (d Date) Before(other Date) bool { return d.t.Before(other.t) }
func (d Date) After(other Date) bool  { return d.t.After(other.t) }
func (d Date) Weekday() time.Weekday  { return d.t.Weekday() }
func (d Date) Day() int               { return d.t.Day() }
func (d Date) Month() time.Month      { return d.t.Month() }
func (d Date) Year() int              { return d.t.Year() }
func (d Date) FirstOfMonth() Date     { return NewDate(d.t.Year(), d.t.Month(), 1) }
func (d Date) Time() time.Time        { return d.t }
