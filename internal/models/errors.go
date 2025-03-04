package models

type JSONErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func PlaceholderGet() (JSONErrorResponse, error) { return JSONErrorResponse{}, nil }
