package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransmissionService_GetActiveDownloads(t *testing.T) {
	t.Run("Empty host returns nothing", func(t *testing.T) {
		s := NewTransmissionService("", "", "")
		downloads, err := s.GetActiveDownloads()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if downloads != nil {
			t.Errorf("expected nil downloads, got %v", downloads)
		}
	})

	t.Run("Successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/transmission/rpc" {
				t.Errorf("expected path /transmission/rpc, got %s", r.URL.Path)
			}
			// Don't check token here as it might be empty if s.Token is updated or if we are not testing token logic here

			resp := transmissionResponse{
				Result: "success",
			}
			resp.Arguments.Torrents = []struct {
				Name   string `json:"name"`
				Status int    `json:"status"`
			}{
				{Name: "Active Torrent", Status: 4},
				{Name: "Paused Torrent", Status: 0},
				{Name: "Seeding Torrent", Status: 6},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		s := NewTransmissionService(server.URL, "user", "pass")
		downloads, err := s.GetActiveDownloads()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(downloads) != 1 || downloads[0] != "Active Torrent" {
			t.Errorf("expected [Active Torrent], got %v", downloads)
		}
	})

	t.Run("Handle 409 Conflict and retry", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.Header().Set("X-Transmission-Session-Id", "new-token")
				w.WriteHeader(http.StatusConflict)
				return
			}

			if r.Header.Get("X-Transmission-Session-Id") != "new-token" {
				t.Errorf("expected retried token to be new-token, got %s", r.Header.Get("X-Transmission-Session-Id"))
			}

			resp := transmissionResponse{
				Result: "success",
			}
			resp.Arguments.Torrents = []struct {
				Name   string `json:"name"`
				Status int    `json:"status"`
			}{
				{Name: "Recovered Torrent", Status: 4},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		s := NewTransmissionService(server.URL, "user", "pass")
		s.Token = "old-token"
		downloads, err := s.GetActiveDownloads()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(downloads) != 1 || downloads[0] != "Recovered Torrent" {
			t.Errorf("expected [Recovered Torrent], got %v", downloads)
		}

		if s.Token != "new-token" {
			t.Errorf("expected service token to be updated to new-token, got %s", s.Token)
		}
	})

	t.Run("RPC error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		s := NewTransmissionService(server.URL, "user", "pass")
		_, err := s.GetActiveDownloads()
		if err == nil {
			t.Error("expected error for non-200 status, got nil")
		}
	})

	t.Run("HTML response instead of JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Error</body></html>"))
		}))
		defer server.Close()

		s := NewTransmissionService(server.URL, "user", "pass")
		_, err := s.GetActiveDownloads()
		if err == nil {
			t.Fatal("expected error for HTML response, got nil")
		}

		expectedError := "expected JSON response, but got text/html"
		if !contains(err.Error(), expectedError) {
			t.Errorf("expected error containing %q, got %v", expectedError, err)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
