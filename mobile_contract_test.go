package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mobileCohort struct {
	id            string
	platform      string
	appVersion    string
	trafficWeight int
}

var mobileCohorts = []mobileCohort{
	{id: "android-legacy-3-2", platform: "android", appVersion: "3.2.0", trafficWeight: 45},
	{id: "android-current-4-1", platform: "android", appVersion: "4.1.0", trafficWeight: 35},
	{id: "ios-current-3-5", platform: "ios", appVersion: "3.5.0", trafficWeight: 15},
	{id: "ios-beta-3-6-rc", platform: "ios", appVersion: "3.6.0-rc1", trafficWeight: 5},
}

func TestMobileCohorts_WeightsSumTo100(t *testing.T) {
	total := 0
	for _, cohort := range mobileCohorts {
		total += cohort.trafficWeight
	}

	if total != 100 {
		t.Fatalf("expected cohort traffic weight total 100, got %d", total)
	}
}

func TestConfigurationContract_MobileCohorts(t *testing.T) {
	InitializeLogger()

	originalLoadConfig := loadConfigFn
	originalGetConfigMap := getConfigMapFn
	t.Cleanup(func() {
		loadConfigFn = originalLoadConfig
		getConfigMapFn = originalGetConfigMap
	})

	loadConfigFn = func() (config, error) {
		return config{
			port:               8080,
			configmapName:      "mobile-config",
			configmapNamespace: "default",
		}, nil
	}

	getConfigMapFn = func(name string, namespace string) (map[string]string, error) {
		if name != "mobile-config" {
			return map[string]string{}, fmt.Errorf("unexpected configmap name: %s", name)
		}
		if namespace != "default" {
			return map[string]string{}, fmt.Errorf("unexpected configmap namespace: %s", namespace)
		}

		return map[string]string{
			"discoveryApiBaseUrl": "http://discovery-api.local",
			"defaultResourceId":   "lebenslagen",
			"resourceTtlSeconds":  "300",
		}, nil
	}

	router := buildRouter()

	for _, cohort := range mobileCohorts {
		t.Run(cohort.id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/configuration", nil)
			req.Header.Set("X-Mobile-Platform", cohort.platform)
			req.Header.Set("X-Mobile-App-Version", cohort.appVersion)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected valid JSON payload, got error: %v", err)
			}

			if body["discoveryApiBaseUrl"] == "" {
				t.Fatal("expected discoveryApiBaseUrl to be present")
			}

			if body["defaultResourceId"] == "" {
				t.Fatal("expected defaultResourceId to be present")
			}
		})
	}
}

func TestConfigurationHandler_Returns500OnConfigMapError(t *testing.T) {
	InitializeLogger()

	originalLoadConfig := loadConfigFn
	originalGetConfigMap := getConfigMapFn
	t.Cleanup(func() {
		loadConfigFn = originalLoadConfig
		getConfigMapFn = originalGetConfigMap
	})

	loadConfigFn = func() (config, error) {
		return config{
			port:               8080,
			configmapName:      "mobile-config",
			configmapNamespace: "default",
		}, nil
	}

	getConfigMapFn = func(name string, namespace string) (map[string]string, error) {
		return map[string]string{}, fmt.Errorf("cluster unavailable")
	}

	req := httptest.NewRequest(http.MethodGet, "/configuration", nil)
	w := httptest.NewRecorder()

	buildRouter().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestIsAliveContract_Returns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/isAlive", nil)
	w := httptest.NewRecorder()

	buildRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
