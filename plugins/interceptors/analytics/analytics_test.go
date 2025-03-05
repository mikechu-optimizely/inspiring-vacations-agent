/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                     *
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

package analytics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optimizely/agent/plugins/interceptors"
)

func TestAnalyticsInterceptor(t *testing.T) {
	// Test that our interceptor is properly registered
	creator, exists := interceptors.Interceptors["analytics"]
	if !exists {
		t.Fatal("Analytics interceptor not registered")
	}

	interceptor := creator()
	if interceptor == nil {
		t.Fatal("Analytics interceptor creator returned nil")
	}

	// Create a test instance of the Analytics interceptor
	analyticsInterceptor, ok := interceptor.(*Analytics)
	if !ok {
		t.Fatal("Failed to cast to Analytics interceptor")
	}

	// Configure the test interceptor
	analyticsInterceptor.Enabled = false // Disable actual GA calls during tests
	analyticsInterceptor.TrackingID = "G-TEST123"

	// Create a simple handler for testing
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response"))
	})

	// Apply our interceptor to the test handler
	handler := analyticsInterceptor.Handler()(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test-path", nil)
	req.Header.Set("User-Agent", "Test User Agent")
	
	// Add a test cookie
	req.AddCookie(&http.Cookie{
		Name:  "_ga",
		Value: "test-client-id",
	})

	// Record the response
	recorder := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Verify the response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, recorder.Code)
	}

	if recorder.Body.String() != "Test response" {
		t.Errorf("Expected response body %q but got %q", "Test response", recorder.Body.String())
	}

	// Note: We don't test the actual GA interaction since it's disabled in tests
	// In a more comprehensive test setup, you would mock the HTTP client
}
