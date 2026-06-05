package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var mesesEspanol = []string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

func FormatFechaEspanol(t time.Time) string {
	return fmt.Sprintf("%d de %s, %d", t.Day(), mesesEspanol[t.Month()-1], t.Year())
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NormalizeCodigo(codigo string) (string, error) {
	codigo = strings.TrimSpace(codigo)

	if codigo == "" {
		return "", fmt.Errorf("código no puede estar vacío")
	}

	if len(codigo) > 4 {
		return "", fmt.Errorf("código no puede tener más de 4 dígitos")
	}

	return fmt.Sprintf("%04s", codigo), nil
}

func FormatCodigo(rawCodigo string) string {
	rawCodigo = strings.TrimSpace(rawCodigo)
	if rawCodigo == "" {
		return ""
	}

	parts := strings.Split(rawCodigo, ".")
	intPart := strings.TrimSpace(parts[0])

	if intPart == "" || intPart == "0" {
		return ""
	}

	numDigits := len(intPart)
	if numDigits > 4 {
		return intPart
	}

	return fmt.Sprintf("%04s", intPart)
}

func CodigoMatchesStored(storedCodigo, inputCodigo string) bool {
	normalizedInput, err := NormalizeCodigo(inputCodigo)
	if err != nil {
		return false
	}

	parts := strings.Split(storedCodigo, ".")
	if len(parts) == 0 {
		return false
	}

	storedWithoutDecimals := strings.TrimRight(parts[0], "0")
	if storedWithoutDecimals == "" {
		storedWithoutDecimals = "0"
	}

	return normalizedInput == fmt.Sprintf("%04s", storedWithoutDecimals)
}

func GenerateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}