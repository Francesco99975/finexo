package models

import (
	"errors"
	"fmt"
	"strings"
)

type SecParams struct {
	Exchange *string `query:"exchange"`
	Country  *string `query:"country"`
	MinPrice *int    `query:"minPrice"`
	MaxPrice *int    `query:"maxPrice"`
	Dividend *bool   `query:"dividend"`
	Order    *string `query:"order"`
	Asc      *string `query:"asc"`
	Limit    *int    `query:"limit"`
}

// Possible valid values for Order and Asc fields
var (
	ValidOrderColumns = map[string]bool{
		"price":     true,
		"volume":    true,
		"avgvolume": true,
		"marketcap": true,
		"pc":        true,
		"ppc":       true,
		"updated":   true,
	}
	ValidAscValues = map[string]bool{
		"asc":  true,
		"desc": true,
	}
)

// Validate method ensures that all fields are valid and normalizes them
func (p *SecParams) Validate() error {
	// Validate and normalize Exchange
	if p.Exchange != nil {
		*p.Exchange = strings.ToUpper(strings.TrimSpace(*p.Exchange))
		if *p.Exchange == "" {
			return errors.New("exchange cannot be an empty string")
		}
	}

	// Validate and normalize Country
	if p.Country != nil {
		*p.Country = strings.ToUpper(strings.TrimSpace(*p.Country))
		if *p.Country == "" {
			return errors.New("country cannot be an empty string")
		}
	}

	// Validate MinPrice (must be non-negative)
	if p.MinPrice != nil && *p.MinPrice < 0 {
		return errors.New("minPrice cannot be negative")
	}

	// Validate MaxPrice (must be non-negative)
	if p.MaxPrice != nil && *p.MaxPrice < 0 {
		return errors.New("maxPrice cannot be negative")
	}

	// Ensure MinPrice is less than or equal to MaxPrice
	if p.MinPrice != nil && p.MaxPrice != nil && *p.MinPrice > *p.MaxPrice {
		return fmt.Errorf("minPrice (%d) cannot be greater than maxPrice (%d)", *p.MinPrice, *p.MaxPrice)
	}

	// Validate and normalize Order
	if p.Order != nil {
		*p.Order = strings.ToLower(strings.TrimSpace(*p.Order))
		if !ValidOrderColumns[*p.Order] {
			return fmt.Errorf("invalid order value: %s, must be one of %v", *p.Order, keys(ValidOrderColumns))
		}
	}

	// Validate and normalize Asc
	if p.Asc != nil {
		*p.Asc = strings.ToLower(strings.TrimSpace(*p.Asc))
		if !ValidAscValues[*p.Asc] {
			return fmt.Errorf("invalid asc value: %s, must be 'asc' or 'desc'", *p.Asc)
		}
	}

	// Validate Limit (must be greater than 0 if provided)
	if p.Limit != nil && *p.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}

	// All validations passed
	return nil
}

// Helper function to get the keys of a map
func keys(m map[string]bool) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}
