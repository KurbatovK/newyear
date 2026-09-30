// Package api реализует HTTP API для расчёта дней до Нового года.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/KurbatovK/newyear/internal/countdown"
)

// DateLayout — формат даты в запросе и ответе (YYYY-MM-DD).
const DateLayout = time.DateOnly

// DaysResponse — успешный ответ.
type DaysResponse struct {
	Date        string `json:"date"`
	NextNewYear string `json:"next_new_year"`
	DaysLeft    int    `json:"days_left"`
}

// ErrorResponse — ответ при ошибке.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse — ответ /healthz.
type HealthResponse struct {
	Status string `json:"status"`
}

// Server хранит зависимости обработчиков.
type Server struct {
	now    func() time.Time
	logger *slog.Logger
}

// NewServer создаёт Server. now задаёт текущее время (nil — time.Now),
// logger может быть nil, тогда запросы не логируются.
func NewServer(now func() time.Time, logger *slog.Logger) *Server {
	if now == nil {
		now = time.Now
	}
	return &Server{now: now, logger: logger}
}

// Handler возвращает обработчик со всеми маршрутами.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/new-year", s.handleDays)
	mux.HandleFunc("GET /healthz", s.handleHealth)

	if s.logger == nil {
		return mux
	}
	return logRequests(s.logger, mux)
}

func (s *Server) handleDays(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	day := s.now()
	if query.Has("date") {
		raw := query.Get("date")
		parsed, err := time.Parse(DateLayout, raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "invalid date " + quote(raw) + ": expected format YYYY-MM-DD",
			})
			return
		}
		day = parsed
	}

	writeJSON(w, http.StatusOK, DaysResponse{
		Date:        day.Format(DateLayout),
		NextNewYear: countdown.NextNewYear(day).Format(DateLayout),
		DaysLeft:    countdown.DaysToNewYear(day),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// статус уже отправлен, сообщить клиенту об ошибке уже нельзя
	_ = json.NewEncoder(w).Encode(v)
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
