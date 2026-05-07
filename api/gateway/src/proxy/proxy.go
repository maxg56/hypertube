package proxy

import (
	"bytes"
	"fmt"
	"gateway/src/middleware"
	"gateway/src/services"
	"gateway/src/utils"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// streamingClient has no timeout — used for video/audio streaming responses.
var streamingClient = &http.Client{}

// defaultClient has a 30-second timeout for all regular requests.
var defaultClient = &http.Client{Timeout: 30 * time.Second}

// ProxyRequest creates a handler that proxies requests to the specified service
func ProxyRequest(serviceName, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		service, exists := services.GetService(serviceName)
		if !exists {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("Service %s not available", serviceName),
			})
			return
		}

		targetURL := service.URL + replacePlaceholders(path, c)
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		req, err := buildRequest(c, targetURL)
		if err != nil {
			log.Printf("Error creating request to %s: %v", targetURL, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		copyHeaders(c, req)

		httpClient := clientForPath(c.Request.URL.Path)
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("Error proxying request to %s: %v", targetURL, err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": fmt.Sprintf("Service %s unavailable", service.Name),
			})
			return
		}
		defer resp.Body.Close()

		copyResponse(c, resp)
	}
}

func buildRequest(c *gin.Context, targetURL string) (*http.Request, error) {
	var body io.Reader
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bodyBytes)
	}
	return http.NewRequest(c.Request.Method, targetURL, body)
}

// clientForPath returns the streaming (no-timeout) client for stream routes,
// and the default client for everything else.
func clientForPath(path string) *http.Client {
	if strings.HasPrefix(path, "/api/v1/stream/") {
		return streamingClient
	}
	return defaultClient
}

// replacePlaceholders replaces path parameters in the target path
func replacePlaceholders(path string, c *gin.Context) string {
	result := path
	for _, param := range c.Params {
		result = strings.ReplaceAll(result, ":"+param.Key, param.Value)
		result = strings.ReplaceAll(result, "*"+param.Key, param.Value)
	}
	return result
}

// copyHeaders copies request headers and adds user context headers
func copyHeaders(c *gin.Context, req *http.Request) {
	for key, values := range c.Request.Header {
		if key == "Host" {
			continue
		}
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	// Override any client-supplied IP headers with the validated TCP-level peer
	// address so backend services can trust X-Real-Ip for rate limiting / audit.
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Real-Ip")
	req.Header.Set("X-Real-Ip", c.RemoteIP())

	if v, ok := c.Get(middleware.CtxUserIDKey); ok {
		if s, ok := v.(string); ok && s != "" {
			req.Header.Set("X-User-ID", s)
			utils.LogDebug("proxy", "User authenticated and ID propagated to service")
		}
	}

	if token := utils.ExtractToken(c); token != "" {
		req.Header.Set("X-JWT-Token", token)
	}
}

// copyResponse copies the upstream response to the client.
// For video/audio content types it streams the body incrementally instead of
// buffering it all in memory, which is required for HTTP Range / 206 responses.
func copyResponse(c *gin.Context, resp *http.Response) {
	contentType := resp.Header.Get("Content-Type")

	c.Status(resp.StatusCode)
	for key, values := range resp.Header {
		for _, v := range values {
			c.Writer.Header().Add(key, v)
		}
	}

	if isStreamingContentType(contentType) {
		streamBody(c, resp)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}
	c.Data(resp.StatusCode, contentType, body)
}

func streamBody(c *gin.Context, resp *http.Response) {
	c.Writer.WriteHeader(resp.StatusCode)
	flusher, canFlush := c.Writer.(http.Flusher)
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n]) //nolint:errcheck
			if canFlush {
				flusher.Flush()
			}
		}
		if readErr != nil {
			break
		}
	}
}

func isStreamingContentType(ct string) bool {
	for _, prefix := range []string{"video/", "audio/", "application/octet-stream"} {
		if strings.HasPrefix(ct, prefix) {
			return true
		}
	}
	return false
}
