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
	MaxWorkers     int
	DefaultLoad    int
}

var Environment *Config

func LoadEnvVariables() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("cannot load environment variables")
	}

	goenv := os.Getenv("GO_ENV")
	var maxWorkers int = 7
	var defaultLoad int = 50
	if goenv == "production" {
		maxWorkers = 3
		defaultLoad = 750
	}

	Environment = &Config{
		Port:           os.Getenv("PORT"),
		Host:           os.Getenv("HOST"),
		GoEnv:          goenv,
		DSN:            os.Getenv("DSN"),
		RapidApiSecret: os.Getenv("RAPID_API_SECRET"),
		MaxWorkers:     maxWorkers,
		DefaultLoad:    defaultLoad,
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
