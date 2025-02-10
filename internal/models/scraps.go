package models

import "time"

type SeekingAlphaScrap struct {
	Holdings   *int       `json:"holdings"`
	FFO        *int       `json:"ffo"`
	PFFO       *int       `json:"pffo"`
	REITiming  *string    `json:"timing"`
	Pr         *int       `json:"pr"`
	Lgr        *int       `json:"lgr"`
	Yog        *int       `json:"yog"`
	Lad        *int       `json:"lad"`
	Frequency  *string    `json:"frequency"`
	ExDivDate  *time.Time `json:"exDivDate"`
	PayoutDate *time.Time `json:"payoutDate"`
}
