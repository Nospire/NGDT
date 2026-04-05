package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

func parseFirstVless(vlessLinks, preferServer string) (uuid, server, port, sni, pbk, sid, fp string, err error) {
	extract := func(line string) (string, string, string, string, string, string, string) {
		parsed, _ := url.Parse(line)
		qs := parsed.Query()
		q := func(key, def string) string {
			if v := qs.Get(key); v != "" {
				return v
			}
			return def
		}
		return parsed.User.Username(), parsed.Hostname(), parsed.Port(),
			q("sni", ""), q("pbk", ""), q("sid", ""), q("fp", "chrome")
	}

	var firstLine string
	for _, line := range strings.Split(vlessLinks, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "vless://") {
			continue
		}
		if firstLine == "" {
			firstLine = line
		}
		if preferServer != "" {
			parsed, e := url.Parse(line)
			if e == nil && strings.Contains(strings.ToUpper(parsed.Fragment), strings.ToUpper(preferServer)) {
				u, srv, p, sni, pbk, sid, fp := extract(line)
				return u, srv, p, sni, pbk, sid, fp, nil
			}
		}
	}
	if firstLine != "" {
		u, srv, p, sni, pbk, sid, fp := extract(firstLine)
		return u, srv, p, sni, pbk, sid, fp, nil
	}
	return "", "", "", "", "", "", "", fmt.Errorf("no vless:// link found")
}

func buildSingBoxConfig(uuid, server, port, sni, pbk, sid, fp string) ([]byte, error) {
	portNum := 0
	fmt.Sscan(port, &portNum)

	cfg := map[string]any{
		"log": map[string]any{"level": "warn"},
		"inbounds": []any{
			map[string]any{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      "127.0.0.1",
				"listen_port": 7890,
			},
		},
		"outbounds": []any{
			map[string]any{
				"type":        "vless",
				"tag":         "proxy",
				"server":      server,
				"server_port": portNum,
				"uuid":        uuid,
				"flow":        "xtls-rprx-vision",
				"tls": map[string]any{
					"enabled":     true,
					"server_name": sni,
					"utls": map[string]any{
						"enabled":     true,
						"fingerprint": fp,
					},
					"reality": map[string]any{
						"enabled":    true,
						"public_key": pbk,
						"short_id":   sid,
					},
				},
			},
			map[string]any{"type": "direct", "tag": "direct"},
		},
		"route": map[string]any{
			"final": "proxy",
		},
	}

	return json.MarshalIndent(cfg, "", "  ")
}

func StartSingBox(vlessLinks, preferServer string) (*os.Process, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(vlessLinks))
	if err == nil {
		vlessLinks = string(decoded)
	}

	exec.Command("pkill", "-f", "sing-box").Run()
	time.Sleep(500 * time.Millisecond)

	uuid, server, port, sni, pbk, sid, fp, err := parseFirstVless(vlessLinks, preferServer)
	if err != nil {
		return nil, fmt.Errorf("parse vless: %w", err)
	}

	cfgJSON, err := buildSingBoxConfig(uuid, server, port, sni, pbk, sid, fp)
	if err != nil {
		return nil, fmt.Errorf("build singbox config: %w", err)
	}

	if err := os.WriteFile(SingBoxConfigPath, cfgJSON, 0600); err != nil {
		return nil, fmt.Errorf("write singbox config: %w", err)
	}

	cmd := exec.Command(SingBoxPath, "run", "-c", SingBoxConfigPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start sing-box: %w", err)
	}

	return cmd.Process, nil
}

func CheckTunnel() (string, error) {
	proxy, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxy)},
		Timeout:   15 * time.Second,
	}
	resp, err := client.Get(TunnelCheckURL)
	if err != nil {
		return "", fmt.Errorf("tunnel check request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Country string `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("tunnel check decode: %w", err)
	}
	return result.Country, nil
}

func StopSingBox(proc *os.Process) {
	if proc != nil {
		_ = proc.Kill()
	}
}
