package models

import "time"

type SEO struct {
	Description string
	Keywords    string
}
type Site struct {
	AppName  string
	Title    string
	Metatags SEO
	Year     int
}

func GetDefaultSite(title string) Site {
	return Site{
		AppName:  "Finexo",
		Title:    title,
		Metatags: SEO{Description: "Compiunt Interest Calulator", Keywords: "tool,finance,calculator,stocks,investment,interest,market"},
		Year:     time.Now().Year(),
	}
}
