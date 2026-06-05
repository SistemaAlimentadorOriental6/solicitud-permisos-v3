package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"solicitud-permisos/internal/config"
)

type Claims struct {
	Codigo string `json:"codigo"`
	Cedula string `json:"cedula"`
	Nombre string `json:"nombre"`
	Area   string `json:"area"`
	jwt.RegisteredClaims
}

func GenerateToken(user struct {
	Codigo string
	Cedula string
	Nombre string
	Area   string
}) (string, error) {
	if config.AppConfig == nil {
		return "", errors.New("configuración no cargada")
	}

	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWT.ExpirationHours) * time.Hour)

	claims := &Claims{
		Codigo: user.Codigo,
		Cedula: user.Cedula,
		Nombre: user.Nombre,
		Area:   user.Area,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "solicitud-permisos",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func ValidateToken(tokenString string) (*Claims, error) {
	if config.AppConfig == nil {
		return nil, errors.New("configuración no cargada")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}