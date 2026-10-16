package admin

import "time"

type memberOut struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	FirstName   string    `json:"first_name"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type updateMemberIn struct {
	Status *string `json:"status"`
	Name   *string `json:"name"`
}
