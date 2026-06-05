package holidays

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Festivo struct {
	Date        string `json:"date"`
	DayOfWeekEs string `json:"day_of_week_es"`
	DayOfWeekEn string `json:"day_of_week_en"`
	DayOfWeekISO int   `json:"day_of_week_iso"`
	NameEs      string `json:"name_es"`
	NameEn      string `json:"name_en"`
}

type festivoResponse struct {
	Data []Festivo `json:"data"`
}

type Service struct {
	apiKey     string
	baseURL    string
	mu         sync.RWMutex
	cache      map[int]map[string]bool
	httpClient *http.Client
}

var instance *Service
var once sync.Once

func InitService(apiKey string) {
	once.Do(func() {
		instance = &Service{
			apiKey:  apiKey,
			baseURL: "https://www.festivos.com.co/api/v1",
			cache:   make(map[int]map[string]bool),
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
		log.Println("Servicio de festivos inicializado")
	})
}

func GetService() *Service {
	return instance
}

func (s *Service) IsHoliday(dateStr string) (bool, error) {
	if s == nil {
		return false, nil
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false, fmt.Errorf("formato de fecha inválido: %s", dateStr)
	}

	year := t.Year()

	yearHolidays, err := s.getYearHolidays(year)
	if err != nil {
		return false, err
	}

	return yearHolidays[dateStr], nil
}

func (s *Service) getYearHolidays(year int) (map[string]bool, error) {
	s.mu.RLock()
	if cached, ok := s.cache[year]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if cached, ok := s.cache[year]; ok {
		return cached, nil
	}

	holidays, err := s.fetchHolidays(year)
	if err != nil {
		return nil, err
	}

	s.cache[year] = holidays
	return holidays, nil
}

func (s *Service) fetchHolidays(year int) (map[string]bool, error) {
	from := fmt.Sprintf("%d-01-01", year)
	to := fmt.Sprintf("%d-12-31", year)
	url := fmt.Sprintf("%s/festivos/range?from=%s&to=%s", s.baseURL, from, to)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error consultando festivos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("API key de festivos inválida")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API de festivos retornó status %d", resp.StatusCode)
	}

	var result festivoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	holidays := make(map[string]bool, len(result.Data))
	for _, f := range result.Data {
		dateStr := strings.TrimSpace(f.Date)
		if dateStr != "" {
			holidays[dateStr] = true
		}
	}

	log.Printf("Festivos cargados para %d: %d dias", year, len(holidays))
	return holidays, nil
}
