package models

import "strings"

// Category хранится кодом, пользователю показывается Label.
type Category string

const (
	CategoryFood          Category = "food"
	CategoryTransport     Category = "transport"
	CategoryHousing       Category = "housing"
	CategoryEntertainment Category = "entertainment"
	CategoryHealth        Category = "health"
	CategoryShopping      Category = "shopping"
	CategoryOther         Category = "other"
)

var Categories = []Category{
	CategoryFood, CategoryTransport, CategoryHousing, CategoryEntertainment,
	CategoryHealth, CategoryShopping, CategoryOther,
}

var categoryLabels = map[Category]string{
	CategoryFood:          "еда",
	CategoryTransport:     "транспорт",
	CategoryHousing:       "жильё",
	CategoryEntertainment: "развлечения",
	CategoryHealth:        "здоровье",
	CategoryShopping:      "покупки",
	CategoryOther:         "другое",
}

func (c Category) Label() string {
	if label, ok := categoryLabels[c]; ok {
		return label
	}

	return string(c)
}

func (c Category) Valid() bool {
	_, ok := categoryLabels[c]
	return ok
}

// ParseCategory принимает код или русское название.
func ParseCategory(s string) (Category, bool) {
	s = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), "ё", "е")
	for _, c := range Categories {
		if s == string(c) || s == strings.ReplaceAll(c.Label(), "ё", "е") {
			return c, true
		}
	}

	return "", false
}
