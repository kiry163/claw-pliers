package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func callAPI(method, path string, payload any) ([]byte, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}

	endpoint, err := base.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload failed: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(data))
		var payload struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &payload) == nil {
			if strings.TrimSpace(payload.Error) != "" {
				msg = payload.Error
			}
		}
		return nil, fmt.Errorf("%s", msg)
	}

	if len(data) == 0 {
		return nil, errors.New("empty response")
	}

	return data, nil
}

func printJSON(data []byte) error {
	if json.Valid(data) {
		var out bytes.Buffer
		if err := json.Indent(&out, data, "", "  "); err == nil {
			fmt.Println(out.String())
			return nil
		}
	}
	fmt.Println(string(data))
	return nil
}
