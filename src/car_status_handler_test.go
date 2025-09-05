package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestTeslaMateAPICarsStatusV1_Integration(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Initialize timezone for tests (normally done in main)
	appUsersTimezone, _ = time.LoadLocation("UTC")

	// Create mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Replace global db with mock
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()

	t.Run("Successful car status retrieval", func(t *testing.T) {
		// Setup mock expectations
		carID := 1
		now := time.Now()

		// Mock car existence check
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(carID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock main status query
		rows := sqlmock.NewRows([]string{
			"id", "name", "model", "trim_badging", "exterior_color", "wheel_type", "spoiler_type", "vin",
			"position_date", "latitude", "longitude", "speed", "power", "odometer", "battery_level",
			"usable_battery_level", "ideal_battery_range_km", "est_battery_range_km", "rated_battery_range_km",
			"outside_temp", "inside_temp", "is_climate_on", "is_preconditioning",
			"state", "state_since", "is_charging", "charging_state",
			"charger_power", "charger_voltage", "charger_phases", "charger_actual_current", "charge_energy_added",
			"unit_of_length", "unit_of_pressure", "unit_of_temperature",
		}).AddRow(
			1, "Test Tesla", "Model 3", "Performance", "Red", "Sport", "None", "5YJ3E1EA4JF123456",
			now, 37.7749, -122.4194, 65, 150, 12345.6, 85,
			83, 400.5, 380.2, 420.8,
			18.5, 22.3, true, false,
			"online", now, true, "charging",
			11000, 240, 3, 45, 5.2,
			"km", "bar", "C",
		)

		mock.ExpectQuery("SELECT.*FROM cars c.*WHERE c.id = \\$1").
			WithArgs(carID).
			WillReturnRows(rows)

		// Setup Gin router and request
		router := gin.New()
		router.GET("/api/v1/cars/:CarID/status", TeslaMateAPICarsStatusV1)

		req, _ := http.NewRequest("GET", "/api/v1/cars/1/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}

		// Check that response body contains expected data
		body := w.Body.String()

		expectedSubstrings := []string{
			`"car_id":1`,
			`"display_name":"Test Tesla"`,
			`"state":"online"`,
			`"model":"Model 3"`,
			`"plugged_in":true`,
			`"charging_state":"charging"`,
			`"charger_power":11000`,
			`"battery_level":85`,
			`"latitude":37.7749`,
			`"longitude":-122.4194`,
		}

		for _, expected := range expectedSubstrings {
			if !contains(body, expected) {
				t.Errorf("Expected response to contain '%s', but it didn't. Body: %s", expected, body)
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Car not found", func(t *testing.T) {
		carID := 999

		// Mock car existence check returning false
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(carID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Setup Gin router and request
		router := gin.New()
		router.GET("/api/v1/cars/:CarID/status", TeslaMateAPICarsStatusV1)

		req, _ := http.NewRequest("GET", "/api/v1/cars/999/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify error response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 (API returns 200 even for errors), got %d", w.Code)
		}

		body := w.Body.String()

		// Check for error in response
		if !contains(body, `"error"`) {
			t.Errorf("Expected error in response, got: %s", body)
		}

		// The handler returns a generic error message, not the specific one
		if !contains(body, "Failed to retrieve car status") {
			t.Errorf("Expected generic error message, got: %s", body)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Invalid car ID", func(t *testing.T) {
		// convertStringToInteger will return 0 for "abc"
		// So we need to mock the query for car ID 0
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(0).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Setup Gin router and request with invalid car ID
		router := gin.New()
		router.GET("/api/v1/cars/:CarID/status", TeslaMateAPICarsStatusV1)

		req, _ := http.NewRequest("GET", "/api/v1/cars/abc/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify error response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 (API returns 200 even for errors), got %d", w.Code)
		}

		body := w.Body.String()

		// Check for error in response (should get a generic error)
		if !contains(body, `"error"`) {
			t.Errorf("Expected error in response, got: %s", body)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Car exists but no position data", func(t *testing.T) {
		carID := 2

		// Mock car existence check  
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(carID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock main status query with mostly null data
		rows := sqlmock.NewRows([]string{
			"id", "name", "model", "trim_badging", "exterior_color", "wheel_type", "spoiler_type", "vin",
			"position_date", "latitude", "longitude", "speed", "power", "odometer", "battery_level",
			"usable_battery_level", "ideal_battery_range_km", "est_battery_range_km", "rated_battery_range_km",
			"outside_temp", "inside_temp", "is_climate_on", "is_preconditioning",
			"state", "state_since", "is_charging", "charging_state",
			"charger_power", "charger_voltage", "charger_phases", "charger_actual_current", "charge_energy_added",
			"unit_of_length", "unit_of_pressure", "unit_of_temperature",
		}).AddRow(
			2, "Minimal Car", "Model Y", nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, false, "disconnected",
			nil, nil, nil, nil, nil,
			"km", "bar", "C",
		)

		mock.ExpectQuery("SELECT.*FROM cars c.*WHERE c.id = \\$1").
			WithArgs(carID).
			WillReturnRows(rows)

		// Setup Gin router and request
		router := gin.New()
		router.GET("/api/v1/cars/:CarID/status", TeslaMateAPICarsStatusV1)

		req, _ := http.NewRequest("GET", "/api/v1/cars/2/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()

		// Verify that handler doesn't crash with null data
		expectedSubstrings := []string{
			`"car_id":2`,
			`"display_name":"Minimal Car"`,
			`"state":"unknown"`, // Should default to unknown when no position data
			`"plugged_in":false`,
			`"charging_state":"disconnected"`,
		}

		for _, expected := range expectedSubstrings {
			if !contains(body, expected) {
				t.Errorf("Expected response to contain '%s', but it didn't. Body: %s", expected, body)
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}