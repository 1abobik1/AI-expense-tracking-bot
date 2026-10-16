package member

type memberEntity struct {
	UserID    int64
	Username  string
	FirstName string
	Name      string
	Status    string
	CreatedAt string
}

func (e *memberEntity) fields() []any {
	return []any{&e.UserID, &e.Username, &e.FirstName, &e.Name, &e.Status, &e.CreatedAt}
}
