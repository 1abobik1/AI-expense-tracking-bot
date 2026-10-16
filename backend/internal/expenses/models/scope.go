package models

// Scope — чьи траты видны; создаётся только в транспорте, нулевое значение не видит ничего.
type Scope struct {
	userID int64
	admin  bool
}

func UserScope(userID int64) Scope { return Scope{userID: userID} }

func AdminScope() Scope { return Scope{admin: true} }

func (s Scope) IsAdmin() bool { return s.admin }

func (s Scope) UserID() int64 { return s.userID }
