package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TransmissionService struct {
	Host     string
	Username string
	Password string
	Token    string
}

func NewTransmissionService(host, user, pass string) *TransmissionService {
	return &TransmissionService{
		Host:     host,
		Username: user,
		Password: pass,
	}
}

type transmissionRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments"`
}

type transmissionResponse struct {
	Arguments struct {
		Torrents []struct {
			Name   string `json:"name"`
			Status int    `json:"status"`
		} `json:"torrents"`
	} `json:"arguments"`
	Result string `json:"result"`
}

func (s *TransmissionService) GetActiveDownloads() ([]string, error) {
	if s.Host == "" {
		return nil, nil
	}

	reqBody := transmissionRequest{
		Method: "torrent-get",
		Arguments: map[string]interface{}{
			"fields": []string{"name", "status"},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/transmission/rpc", s.Host)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if s.Token != "" {
		req.Header.Set("X-Transmission-Session-Id", s.Token)
	}

	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle 409 Conflict which is common in Transmission if token is missing/expired
	if resp.StatusCode == http.StatusConflict {
		newToken := resp.Header.Get("X-Transmission-Session-Id")
		if newToken != "" {
			s.Token = newToken
			// Retry once
			return s.GetActiveDownloads()
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transmission rpc error: %s", resp.Status)
	}

	// Check if the response is actually JSON
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		// Log or return a better error if we get HTML
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("expected JSON response, but got %s: %s", contentType, string(body))
	}

	var transResp transmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transResp); err != nil {
		return nil, err
	}

	var activeNames []string
	for _, t := range transResp.Arguments.Torrents {
		// status: 4 = downloading, 6 = seeding (if we want to show both)
		// Usually active means downloading (4)
		if t.Status == 4 {
			activeNames = append(activeNames, t.Name)
		}
	}

	return activeNames, nil
}
