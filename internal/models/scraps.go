package models

type SeekingAlphaScrap struct {
	Holdings   *int    `json:"holdings"`
	FFO        *int    `json:"ffo"`
	PFFO       *int    `json:"pffo"`
	REITiming  *string `json:"timing"`
	Pr         *int    `json:"pr"`
	Lgr        *int    `json:"lgr"`
	Yog        *int    `json:"yog"`
	Lad        *int    `json:"lad"`
	Frequency  *string `json:"frequency"`
	ExDivDate  *string `json:"exDivDate"`
	PayoutDate *string `json:"payoutDate"`
}
