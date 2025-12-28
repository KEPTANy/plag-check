package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ReverseProxy struct {
	userServiceURL        string
	fileStorageServiceURL string
	analysisServiceURL    string
}

func NewReverseProxy(userServiceURL, fileStorageServiceURL, analysisServiceURL string) *ReverseProxy {
	return &ReverseProxy{
		userServiceURL:        userServiceURL,
		fileStorageServiceURL: fileStorageServiceURL,
		analysisServiceURL:    analysisServiceURL,
	}
}

func (p *ReverseProxy) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	targetURL := p.getTargetURL(r.URL.Path)
	if targetURL == "" {
		http.Error(w, `{"error": "route not found"}`, http.StatusNotFound)
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Failed to parse target URL: %v", err)
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	proxyURL := *target
	proxyURL.Path = r.URL.Path
	proxyURL.RawQuery = r.URL.RawQuery

	req, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		if strings.ToLower(key) == "host" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to proxy request: %v", err)
		http.Error(w, `{"error": "service unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if resp.Body != nil {
		io.Copy(w, resp.Body)
	}
}

func (p *ReverseProxy) getTargetURL(path string) string {
	if strings.HasPrefix(path, "/auth/") {
		return p.userServiceURL
	}
	if strings.HasPrefix(path, "/files/") {
		return p.fileStorageServiceURL
	}
	if strings.HasPrefix(path, "/analysis/") {
		return p.analysisServiceURL
	}
	return ""
}
