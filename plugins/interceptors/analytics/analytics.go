/****************************************************************************
 * Copyright 2025, Inspiring Vacations and contributors                     *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package analytics implements an interceptor for tracking API usage with Google Analytics
package analytics

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/plugins/interceptors"
)

// Analytics implements the Interceptor plugin interface for Google Analytics tracking
type Analytics struct {
	// Configuration fields
	TrackingID string // Google Analytics tracking ID (e.g., UA-XXXXX-Y or G-XXXXXXX)
	Enabled    bool   // Whether analytics tracking is enabled
	EndpointURL string // Google Analytics endpoint URL (defaults to GA4 endpoint)
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader captures the status code and calls the original WriteHeader
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body and calls the original Write
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Handler returns a middleware function that tracks API usage with Google Analytics
func (a *Analytics) Handler() func(http.Handler) http.Handler {
	// Default endpoint for GA4
	if a.EndpointURL == "" {
		a.EndpointURL = "https://www.google-analytics.com/mp/collect"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if analytics is disabled
			if !a.Enabled || a.TrackingID == "" {
				next.ServeHTTP(w, r)
				return
			}

			startTime := time.Now()

			// Create a wrapper for the response writer to capture response details
			responseBuffer := &bytes.Buffer{}
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code
				body:           responseBuffer,
			}

			// Create a copy of the request body for analysis
			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = ioutil.ReadAll(r.Body)
				// Restore the request body for the next handlers
				r.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Continue with the normal request handling
			next.ServeHTTP(wrappedWriter, r)

			// Calculate request duration
			duration := time.Since(startTime).Milliseconds()

			// Prepare analytics data to send to Google Analytics
			// This is a simplified version - adjust to your needs
			eventData := map[string]interface{}{
				"client_id": getClientID(r),
				"events": []map[string]interface{}{
					{
						"name": "api_request",
						"params": map[string]interface{}{
							"path":             r.URL.Path,
							"method":           r.Method,
							"status_code":      wrappedWriter.statusCode,
							"response_time_ms": duration,
							"user_agent":       r.UserAgent(),
							"ip_address":       getIPAddress(r),
						},
					},
				},
			}

			// Send data to Google Analytics in a separate goroutine to not block the response
			go a.sendToGA(eventData)

			log.Info().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Int("status", wrappedWriter.statusCode).
				Int64("duration_ms", duration).
				Msg("Analytics tracking sent")
		})
	}
}

// sendToGA sends event data to Google Analytics
func (a *Analytics) sendToGA(eventData map[string]interface{}) {
	// Prepare the URL with the tracking ID
	url := a.EndpointURL + "?measurement_id=" + a.TrackingID + "&api_secret=YOUR_API_SECRET" // You would need to set this in config

	// Convert event data to JSON
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal analytics data")
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create analytics request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send analytics data")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Error().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("Analytics request failed")
	}
}

// getClientID extracts a client ID from the request
// In a real implementation, you might use cookies or other identifiers
func getClientID(r *http.Request) string {
	// Use a cookie, header, or session ID as the client identifier
	cookie, err := r.Cookie("_ga")
	if err == nil && cookie != nil {
		return cookie.Value
	}

	// Fallback to IP + User-Agent hash if no cookie exists
	// In a real implementation, you would generate a proper UUID
	return getIPAddress(r) + r.UserAgent()
}

// getIPAddress extracts the client IP address from the request
func getIPAddress(r *http.Request) string {
	// Try common headers for IP addresses
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if ip := r.Header.Get(header); ip != "" {
			// Take the first IP if it's a comma-separated list
			return strings.Split(ip, ",")[0]
		}
	}
	// Fallback to remote address
	return strings.Split(r.RemoteAddr, ":")[0]
}

// Register our interceptor as "analytics"
func init() {
	interceptors.Add("analytics", func() interceptors.Interceptor {
		return &Analytics{}
	})
}
