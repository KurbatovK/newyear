package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, time.September, 30, 15, 4, 5, 0, time.UTC)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(NewServer(func() time.Time { return fixedNow }, logger).Handler())
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return resp, body
}

func TestDaysEndpoint(t *testing.T) {
	srv := newTestServer(t)

	tests := []struct {
		name  string
		query string
		want  DaysResponse
	}{
		{
			name:  "current date",
			query: "",
			want:  DaysResponse{Date: "2026-09-30", NextNewYear: "2027-01-01", DaysLeft: 93},
		},
		{
			name:  "end of year",
			query: "?date=2025-12-31",
			want:  DaysResponse{Date: "2025-12-31", NextNewYear: "2026-01-01", DaysLeft: 1},
		},
		{
			name:  "start of leap year",
			query: "?date=2024-01-01",
			want:  DaysResponse{Date: "2024-01-01", NextNewYear: "2025-01-01", DaysLeft: 366},
		},
		{
			name:  "Feb 29",
			query: "?date=2024-02-29",
			want:  DaysResponse{Date: "2024-02-29", NextNewYear: "2025-01-01", DaysLeft: 307},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := get(t, srv.URL+"/api/v1/new-year"+tt.query)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d; body: %s", resp.StatusCode, http.StatusOK, body)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			var got DaysResponse
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("decoding response %s: %v", body, err)
			}
			if got != tt.want {
				t.Errorf("response = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDaysEndpointInvalidDate(t *testing.T) {
	srv := newTestServer(t)

	for _, q := range []string{
		"?date=",
		"?date=abc",
		"?date=2025-02-29", // не високосный
		"?date=2025-13-01",
		"?date=31.12.2025",
		"?date=2025-12-31T10:00:00Z",
	} {
		t.Run(q, func(t *testing.T) {
			resp, body := get(t, srv.URL+"/api/v1/new-year"+q)

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body: %s", resp.StatusCode, http.StatusBadRequest, body)
			}
			var got ErrorResponse
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("decoding response %s: %v", body, err)
			}
			if got.Error == "" {
				t.Error("error message is empty")
			}
		})
	}
}

func TestDaysEndpointMethodNotAllowed(t *testing.T) {
	srv := newTestServer(t)

	resp, err := http.Post(srv.URL+"/api/v1/new-year", "application/json", bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestUnknownPath(t *testing.T) {
	srv := newTestServer(t)

	resp, _ := get(t, srv.URL+"/nope")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHealth(t *testing.T) {
	srv := newTestServer(t)

	resp, body := get(t, srv.URL+"/healthz")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got HealthResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decoding response %s: %v", body, err)
	}
	if got.Status != "ok" {
		t.Errorf("status field = %q, want %q", got.Status, "ok")
	}
}
