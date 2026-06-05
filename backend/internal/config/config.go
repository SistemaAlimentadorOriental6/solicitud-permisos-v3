package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
}

type Config struct {
	Server            ServerConfig
	Database          DatabaseConfig
	UnoEE             DatabaseConfig
	SolicitudPermisos DatabaseConfig
	JWT               JWTConfig
	S3                S3Config
	FestivosAPIKey    string
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

var AppConfig *Config

// LoadConfig carga la configuración desde el archivo .env
func LoadConfig(ruta string) (*Config, error) {
	if err := godotenv.Load(ruta); err != nil {
		log.Println("Archivo .env no encontrado, usando variables de entorno del sistema")
	}

	// Validar variables obligatorias
	variablesRequeridas := []string{
		"DB_HOST", "DB_DATABASE", "DB_USER", "DB_PASSWORD",
		"UNOEE_HOST", "UNOEE_DATABASE", "UNOEE_USER", "UNOEE_PASSWORD",
		"JWT_SECRET",
	}

	for _, variable := range variablesRequeridas {
		if os.Getenv(variable) == "" {
			return nil, fmt.Errorf("variable de entorno requerida '%s' no está definida en el .env", variable)
		}
	}

	puerto, _ := strconv.Atoi(obtenerEnv("SERVER_PORT", "8080"))
	puertoDb, _ := strconv.Atoi(obtenerEnv("DB_PORT", "1433"))
	puertoUnoee, _ := strconv.Atoi(obtenerEnv("UNOEE_PORT", "1433"))
	expJwt, _ := strconv.Atoi(obtenerEnv("JWT_EXPIRATION_HOURS", "24"))

	spPort, _ := strconv.Atoi(obtenerEnv("SP_PORT", "3306"))

	cfg := &Config{
		Server: ServerConfig{
			Host: obtenerEnv("SERVER_HOST", "0.0.0.0"),
			Port: puerto,
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     puertoDb,
			Database: os.Getenv("DB_DATABASE"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
		UnoEE: DatabaseConfig{
			Host:     os.Getenv("UNOEE_HOST"),
			Port:     puertoUnoee,
			Database: os.Getenv("UNOEE_DATABASE"),
			User:     os.Getenv("UNOEE_USER"),
			Password: os.Getenv("UNOEE_PASSWORD"),
		},
		SolicitudPermisos: DatabaseConfig{
			Host:     os.Getenv("DBPERMISOS_HOST"),
			Port:     spPort,
			Database: os.Getenv("DBPERMISOS_NAME"),
			User:     os.Getenv("DBPERMISOS_USER"),
			Password: os.Getenv("DBPERMISOS_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:          os.Getenv("JWT_SECRET"),
			ExpirationHours: expJwt,
		},
		S3: S3Config{
			AccessKeyID:     strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID")),
			SecretAccessKey: strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY")),
			BucketName:      strings.TrimSpace(os.Getenv("AWS_STORAGE_BUCKET_NAME")),
			Region:          strings.TrimSpace(os.Getenv("AWS_S3_REGION_NAME")),
		},
		FestivosAPIKey: os.Getenv("FESTIVOS_API_KEY"),
	}

	AppConfig = cfg
	return cfg, nil
}

func obtenerEnv(clave, valorDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return valorDefecto
}
