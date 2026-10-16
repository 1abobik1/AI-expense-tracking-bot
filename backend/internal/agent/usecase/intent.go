package usecase

import (
	"encoding/json"
	"strings"
)

// Intent намеренно без поля пользователя: модель не решает, чьи это траты.
type Intent struct {
	Action   string    `json:"action"` // add | list | update | delete | help
	Items    []NewItem `json:"items"`  // add
	Period   string    `json:"period"` // list: см. periods
	From     string    `json:"from"`   // list, period=custom
	To       string    `json:"to"`
	Category string    `json:"category"` // list: фильтр
	Target   Target    `json:"target"`   // update, delete: какую трату
	Changes  Changes   `json:"changes"`  // update
}

type NewItem struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
	Comment  string  `json:"comment"`
}

type Target struct {
	Last   bool    `json:"last"`
	Text   string  `json:"text"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

type Changes struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
	Comment  string  `json:"comment"`
}

const (
	actionAdd    = "add"
	actionList   = "list"
	actionUpdate = "update"
	actionDelete = "delete"
	actionHelp   = "help"
)

// parseIntent терпит обёртку ```json и текст вокруг.
func parseIntent(text string) (Intent, bool) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	}

	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return Intent{}, false
	}

	var in Intent
	if err := json.Unmarshal([]byte(text[start:end+1]), &in); err != nil {
		return Intent{}, false
	}

	in.Action = strings.ToLower(strings.TrimSpace(in.Action))

	return in, in.Action != ""
}
