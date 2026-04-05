package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type startSessionResponse struct {
	SessionID    string `json:"session_id"`
	MihomoConfig string `json:"mihomo_config"`
}

func StartSession() (sessionID, mihomoConfig string, err error) {
	body, _ := json.Marshal(map[string]string{"client_type": "ngdt"})
	resp, err := http.Post(OrchestratorURL+"/api/session/start", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("start session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("start session: unexpected status %d", resp.StatusCode)
	}

	var result startSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("start session decode: %w", err)
	}
	return result.SessionID, result.MihomoConfig, nil
}

func SendHeartbeat(sessionID string) error {
	resp, err := http.Post(OrchestratorURL+"/api/session/"+sessionID+"/heartbeat", "application/json", nil)
	if err != nil {
		return fmt.Errorf("heartbeat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func CompleteSession(sessionID, serverUsed, countryDetected string, success bool) error {
	body, _ := json.Marshal(map[string]any{
		"server_used":      serverUsed,
		"country_detected": countryDetected,
		"success":          success,
	})
	resp, err := http.Post(OrchestratorURL+"/api/session/"+sessionID+"/complete", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("complete session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("complete session: unexpected status %d", resp.StatusCode)
	}
	return nil
}
