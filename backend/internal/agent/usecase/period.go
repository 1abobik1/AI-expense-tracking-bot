package usecase

import (
	"time"

	"expensebot/internal/expenses/models"
)

// Модель выбирает только название периода, даты считает код.
const (
	periodToday     = "today"
	periodYesterday = "yesterday"
	periodThisWeek  = "this_week"
	periodLast7Days = "last_7_days"
	periodThisMonth = "this_month"
	periodLastMonth = "last_month"
	periodThisYear  = "this_year"
	periodAll       = "all"
	periodCustom    = "custom"
)

func resolvePeriod(in Intent, today models.Date) (from, to models.Date, label string) {
	switch in.Period {
	case periodToday:
		return today, today, "сегодня"
	case periodYesterday:
		y := today.AddDays(-1)
		return y, y, "вчера"
	case periodThisWeek:
		offset := (int(today.Weekday()) + 6) % 7 // понедельник — начало недели
		return today.AddDays(-offset), today, "на этой неделе"
	case periodLast7Days:
		return today.AddDays(-6), today, "за последние 7 дней"
	case periodLastMonth:
		first := today.FirstOfMonth().AddMonths(-1)
		return first, today.FirstOfMonth().AddDays(-1), "за " + monthName(first.Month(), false)
	case periodThisYear:
		return models.NewDate(today.Year(), time.January, 1), today, "в этом году"
	case periodAll:
		return models.Date{}, models.Date{}, "за всё время"
	case periodCustom:
		from, errFrom := models.ParseDate(in.From)

		to, errTo := models.ParseDate(in.To)
		if errFrom == nil || errTo == nil {
			return from, to, "с " + orDash(from) + " по " + orDash(to)
		}
	}

	return today.FirstOfMonth(), today, "в этом месяце"
}

func orDash(d models.Date) string {
	if d.IsZero() {
		return "…"
	}

	return shortDate(d)
}

var months = [...]struct{ nominative, genitive string }{
	{"январь", "января"},
	{"февраль", "февраля"},
	{"март", "марта"},
	{"апрель", "апреля"},
	{"май", "мая"},
	{"июнь", "июня"},
	{"июль", "июля"},
	{"август", "августа"},
	{"сентябрь", "сентября"},
	{"октябрь", "октября"},
	{"ноябрь", "ноября"},
	{"декабрь", "декабря"},
}

func monthName(m time.Month, genitive bool) string {
	if genitive {
		return months[m-1].genitive
	}

	return months[m-1].nominative
}

func shortDate(d models.Date) string {
	return itoa(d.Day()) + " " + monthName(d.Month(), true)
}

func humanDate(d, today models.Date) string {
	switch d {
	case today:
		return "сегодня"
	case today.AddDays(-1):
		return "вчера"
	}

	return shortDate(d)
}
