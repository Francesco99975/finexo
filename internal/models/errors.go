package models

type JSONErrorResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

func PlaceholderGet() (JSONErrorResponse, error) { return JSONErrorResponse{}, nil }
