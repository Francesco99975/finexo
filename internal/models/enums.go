package models

import (
	"errors"
	"strings"
)

type Timing string
type Frequency string

const (
	TimingFWD Timing = "fwd"
	TimingTTM Timing = "ttm"

	FrequencyUnknown   Frequency = "unknown"
	FrequencyWeekly    Frequency = "weekly"
	FrequencyBiweekly  Frequency = "biweekly"
	FrequencyMonthly   Frequency = "monthly"
	FrequencyQuarterly Frequency = "quarterly"
	FrequencySemi      Frequency = "semi-annual"
	FrequencyYearly    Frequency = "annual"
)

func ParseFrequency(frequency string) (Frequency, error) {
	switch strings.ToLower(frequency) {
	case "weekly":
		return FrequencyWeekly, nil
	case "biweekly":
		return FrequencyBiweekly, nil
	case "monthly":
		return FrequencyMonthly, nil
	case "quarterly":
		return FrequencyQuarterly, nil
	case "semi-annual":
		return FrequencySemi, nil
	case "annual":
		return FrequencyYearly, nil
	default:
		return FrequencyUnknown, errors.New("invalid frequency")
	}
}
