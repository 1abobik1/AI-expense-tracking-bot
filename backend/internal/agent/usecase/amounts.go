package usecase

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// normalizeAmounts: «пятихатка» → 500, «2 косаря» → 2000; слова заменяются, только если в тексте нет цифр.
func normalizeAmounts(text string) string {
	p := amountPatterns()
	text = strings.ReplaceAll(strings.ToLower(text), "ё", "е")
	hadDigits := strings.ContainsAny(text, "0123456789")

	text = p.groupedNumber.ReplaceAllStringFunc(text, func(s string) string { return strings.ReplaceAll(s, " ", "") })
	text = p.decimalComma.ReplaceAllString(text, "$1.$2")
	text = replaceWords(text, p.halfThousand, func([]string) string { return "1500" })

	text = replaceWords(text, p.thousands, func(m []string) string {
		n, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", "."), 64)
		if err != nil {
			return m[0]
		}

		return strconv.FormatFloat(n*1000, 'f', -1, 64)
	})
	if hadDigits {
		return text
	}

	for _, s := range p.slang {
		text = replaceWords(text, s.re, func([]string) string { return s.amount })
	}

	return text
}

// \b в Go понимает только латиницу.
const (
	wordStart = `(?:^|[^\p{L}\d])`
	wordEnd   = `(?:$|[^\p{L}])`
)

func word(pattern string) *regexp.Regexp {
	return regexp.MustCompile(wordStart + `(` + pattern + `)` + wordEnd)
}

type slangAmount struct {
	re     *regexp.Regexp
	amount string
}

type patterns struct {
	groupedNumber, decimalComma, halfThousand, thousands *regexp.Regexp
	slang                                                []slangAmount
}

// Лениво: регулярки с \p{L} нужны только агенту, а на старте они стоили CPU каждому запросу Worker.
var amountPatterns = sync.OnceValue(func() patterns {
	return patterns{
		groupedNumber: regexp.MustCompile(`\b\d{1,3}(?: \d{3})+\b`),
		// Слабые модели читают «99,90» как 990.
		decimalComma: regexp.MustCompile(`(\d),(\d{1,2})\b`),
		halfThousand: word(`полтор[аы]\s+(?:косар\p{L}*|касар\p{L}*|тысяч\p{L}*|тыщ\p{L}*|штук\p{L}*)`),
		// «к» только слитно («350 к обеду» — не тысячи); «3 штуки» — количество, не множитель.
		thousands: word(`(\d+(?:[.,]\d+)?)(?:[кk]|\s*(?:тыс\.?|тысяч\p{L}*|тыщ\p{L}*|косар\p{L}*|касар\p{L}*))`),
		// Порядок важен: «полтыщи» раньше «тыщи».
		slang: []slangAmount{
			{word(`полтыщ\p{L}*|полтысячи`), "500"},
			{word(`полтор[аы]`), "1500"},
			{word(`пятихат\p{L}*|пятисотк\p{L}*`), "500"},
			// Только формы без числа перед ними: «косаря», «сотки» идут после числа, их разбирает модель.
			{word(`косарь|косарик\p{L}*|касарь|касарик\p{L}*|тыщ[ау]|тысяч[ау]|штук[ау]`), "1000"},
			{word(`сотк[ау]|стольник\p{L}*`), "100"},
			{word(`полтинник\p{L}*|полтос\p{L}*`), "50"},
			{word(`червон\p{L}*`), "10"},
			{word(`двушк[аиуе]`), "2000"},
			{word(`трешк[аиуе]`), "3000"},
			{word(`пятер[аыу]|пятерк[аиуе]`), "5000"},
			{word(`шестер[аыу]`), "6000"},
			{word(`десятк[аиуе]|десятер[аыу]`), "10000"},
		},
	}
})

// replaceWords продолжает поиск с конца слова, а не совпадения: разделитель нужен и следующему слову.
func replaceWords(text string, re *regexp.Regexp, repl func(groups []string) string) string {
	var b strings.Builder

	pos := 0
	for pos <= len(text) {
		loc := re.FindStringSubmatchIndex(text[pos:])
		if loc == nil {
			break
		}

		groups := make([]string, 0, len(loc)/2-1)
		for i := 2; i < len(loc); i += 2 {
			if loc[i] < 0 {
				groups = append(groups, "")
				continue
			}

			groups = append(groups, text[pos+loc[i]:pos+loc[i+1]])
		}

		b.WriteString(text[pos : pos+loc[2]])
		b.WriteString(repl(groups))

		pos += loc[3]
	}

	if pos < len(text) {
		b.WriteString(text[pos:])
	}

	return b.String()
}
