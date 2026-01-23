package server

import (
	"io"
	"net/http"
	"strings"
	"time"
)

func detectPublicIP() (string, error) {
	c := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := c.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}
