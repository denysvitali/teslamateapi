package main

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCarStatusService_GetCarStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	service := NewCarStatusService(db)

	t.Run("Car exists and has data", func(t *testing.T) {
		carID := 1
		now := time.Now()

		// Mock car existence check
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(carID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Mock main status query with all expected columns
		rows := sqlmock.NewRows([]string{
			"id", "name", "model", "trim_badging", "exterior_color", "wheel_type", "spoiler_type", "vin",
			"position_date", "latitude", "longitude", "speed", "power", "odometer", "battery_level",
			"usable_battery_level", "ideal_battery_range_km", "est_battery_range_km", "rated_battery_range_km",
			"outside_temp", "inside_temp", "is_climate_on",
			"state", "state_since", "is_charging", "charging_state",
			"charger_power", "charger_voltage", "charger_phases", "charger_actual_current", "charge_energy_added",
			"unit_of_length", "unit_of_pressure", "unit_of_temperature",
		}).AddRow(
			1, "Test Car", "Model 3", "Performance", "Red", "Sport", "None", "5YJ3E1EA4JF123456",
			now, 37.7749, -122.4194, 65, 150, 12345.6, 85,
			83, 400.5, 380.2, 420.8,
			18.5, 22.3, true,
			"online", now, true, "charging",
			11000, 240, 3, 45, 5.2,
			"km", "bar", "C",
		)

		mock.ExpectQuery("SELECT.*FROM cars c.*WHERE c.id = \\$1").
			WithArgs(carID).
			WillReturnRows(rows)

		result, err := service.GetCarStatus(carID)
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Verify key fields
		if result.CarID != 1 {
			t.Errorf("Expected CarID 1, got %d", result.CarID)
		}

		if result.Name.String != "Test Car" {
			t.Errorf("Expected name 'Test Car', got %s", result.Name.String)
		}

		if !result.IsCharging.Bool {
			t.Error("Expected car to be charging")
		}

		if result.ChargingState.String != "charging" {
			t.Errorf("Expected charging state 'charging', got %s", result.ChargingState.String)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Car does not exist", func(t *testing.T) {
		carID := 999

		// Mock car existence check returning false
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(carID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		result, err := service.GetCarStatus(carID)

		if err == nil {
			t.Error("Expected error for non-existent car, got nil")
		}

		if result != nil {
			t.Error("Expected nil result for non-existent car")
		}

		expectedError := "car with ID 999 does not exist"
		if err.Error() != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Database error on existence check", func(t *testing.T) {
		carID := 1
		expectedErr := errors.New("database connection failed")

		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM cars WHERE id=\\$1\\)").
			WithArgs(carID).
			WillReturnError(expectedErr)

		result, err := service.GetCarStatus(carID)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if result != nil {
			t.Error("Expected nil result on database error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCarStatusService_DetermineVehicleState(t *testing.T) {
	service := NewCarStatusService(nil) // No DB needed for this test

	t.Run("State from database", func(t *testing.T) {
		data := &CarStatusData{
			State: sql.NullString{String: "asleep", Valid: true},
		}

		state := service.DetermineVehicleState(data)

		if state != "asleep" {
			t.Errorf("Expected state 'asleep', got '%s'", state)
		}
	})

	t.Run("State from position timestamp - online", func(t *testing.T) {
		recentTime := time.Now().Add(-2 * time.Minute)
		data := &CarStatusData{
			State:        sql.NullString{Valid: false},
			PositionDate: sql.NullTime{Time: recentTime, Valid: true},
		}

		state := service.DetermineVehicleState(data)

		if state != "online" {
			t.Errorf("Expected state 'online', got '%s'", state)
		}
	})

	t.Run("State from position timestamp - asleep", func(t *testing.T) {
		oldTime := time.Now().Add(-15 * time.Minute)
		data := &CarStatusData{
			State:        sql.NullString{Valid: false},
			PositionDate: sql.NullTime{Time: oldTime, Valid: true},
		}

		state := service.DetermineVehicleState(data)

		if state != "asleep" {
			t.Errorf("Expected state 'asleep', got '%s'", state)
		}
	})

	t.Run("State from position timestamp - offline", func(t *testing.T) {
		veryOldTime := time.Now().Add(-45 * time.Minute)
		data := &CarStatusData{
			State:        sql.NullString{Valid: false},
			PositionDate: sql.NullTime{Time: veryOldTime, Valid: true},
		}

		state := service.DetermineVehicleState(data)

		if state != "offline" {
			t.Errorf("Expected state 'offline', got '%s'", state)
		}
	})

	t.Run("No state data available", func(t *testing.T) {
		data := &CarStatusData{
			State:        sql.NullString{Valid: false},
			PositionDate: sql.NullTime{Valid: false},
		}

		state := service.DetermineVehicleState(data)

		if state != "unknown" {
			t.Errorf("Expected state 'unknown', got '%s'", state)
		}
	})
}