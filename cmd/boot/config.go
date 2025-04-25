package boot

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	Host           string
	GoEnv          string
	DSN            string
	RapidApiSecret string
}

var Environment *Config

func LoadEnvVariables() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("cannot load environment variables")
	}

	Environment = &Config{
		Port:           os.Getenv("PORT"),
		Host:           os.Getenv("HOST"),
		GoEnv:          os.Getenv("GO_ENV"),
		DSN:            os.Getenv("DSN"),
		RapidApiSecret: os.Getenv("RAPID_API_SECRET"),
	}

	return err
}

type PlanLimits struct {
	AllowedParams []string
	MaxParams     int
}

var PlanConfigs = map[string]PlanLimits{
	"free": {
		AllowedParams: []string{"exchange", "country", "currency"},
		MaxParams:     2,
	},
	"basic": {
		AllowedParams: []string{"exchange", "country", "currency"},
		MaxParams:     5,
	},
	"pro": {
		AllowedParams: nil, // nil means all parameters allowed
		MaxParams:     -1,  // -1 means no limit
	},
}
