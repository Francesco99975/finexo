package boot

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port  string
	Host  string
	GoEnv string
	DSN   string
}

var Environment = &Config{}

func LoadEnvVariables() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("cannot load environment variables")
	}

	Environment.Port = os.Getenv("PORT")
	Environment.Host = os.Getenv("HOST")
	Environment.GoEnv = os.Getenv("GO_ENV")
	Environment.DSN = os.Getenv("DSN")

	return err
}
