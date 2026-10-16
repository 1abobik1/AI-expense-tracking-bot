package models

import (
	"math"
	"strconv"
	"strings"
)

// Money — сумма в копейках.
type Money int64

// MaxMoney отсекает опечатки и мусор от ИИ.
const MaxMoney Money = 100_000_000_00

func MoneyFromRubles(rubles float64) Money {
	return Money(math.Round(rubles * 100))
}

func (m Money) Rubles() float64 {
	return float64(m) / 100
}

// String — «1 200 ₽», «99,90 ₽».
func (m Money) String() string {
	sign := ""
	if m < 0 {
		sign, m = "-", -m
	}

	rubles := groupThousands(strconv.FormatInt(int64(m)/100, 10))
	if kopecks := int64(m) % 100; kopecks != 0 {
		return sign + rubles + "," + leftPad2(kopecks) + " ₽"
	}

	return sign + rubles + " ₽"
}

func groupThousands(digits string) string {
	var b strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteRune(' ')
		}

		b.WriteRune(r)
	}

	return b.String()
}

func leftPad2(n int64) string {
	if n < 10 {
		return "0" + strconv.FormatInt(n, 10)
	}

	return strconv.FormatInt(n, 10)
}
