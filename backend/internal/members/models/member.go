package models

import (
	"strconv"
	"time"
)

// Status: новые участники ждут одобрения главы семьи, до него бот не пускает их к ИИ.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

func (s Status) Valid() bool {
	return s == StatusPending || s == StatusApproved || s == StatusRejected
}

const MaxNameLength = 50

type Member struct {
	UserID    int64
	Username  string // из Telegram
	FirstName string // из Telegram
	Name      string // задаёт глава семьи: «Мама», «Сын»
	Status    Status
	CreatedAt time.Time
}

func (m Member) DisplayName() string {
	switch {
	case m.Name != "":
		return m.Name
	case m.FirstName != "" && m.Username != "":
		return m.FirstName + " (@" + m.Username + ")"
	case m.FirstName != "":
		return m.FirstName
	case m.Username != "":
		return "@" + m.Username
	}

	return "id " + strconv.FormatInt(m.UserID, 10)
}

// UpdateMemberIn: nil-поля не меняются.
type UpdateMemberIn struct {
	Status *Status
	Name   *string
}
