package helpers

import (
	"unicode"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func Capitalize(s string) string {
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

func FormatPrice(price float64) string {
	p := message.NewPrinter(language.English)
	cur := currency.CAD
	return p.Sprintf("%v", cur.Amount(price))
}
