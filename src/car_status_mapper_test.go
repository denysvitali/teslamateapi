package main

import (
	"database/sql"
	"testing"
	"time"
)

func TestCarStatusMapper_MapToResponse(t *testing.T) {
	// Initialize timezone for tests (normally done in main)
	appUsersTimezone, _ = time.LoadLocation("UTC")
	
	mapper := NewCarStatusMapper()

	t.Run("Complete data mapping", func(t *testing.T) {
		now := time.Now()
		data := &CarStatusData{
			CarID:             1,
			Name:              sql.NullString{String: "Test Car", Valid: true},
			Model:             sql.NullString{String: "Model 3", Valid: true},
			TrimBadging:       sql.NullString{String: "Performance", Valid: true},
			ExteriorColor:     sql.NullString{String: "Red", Valid: true},
			WheelType:         sql.NullString{String: "Sport", Valid: true},
			SpoilerType:       sql.NullString{String: "None", Valid: true},
			Latitude:          sql.NullFloat64{Float64: 37.7749, Valid: true},
			Longitude:         sql.NullFloat64{Float64: -122.4194, Valid: true},
			Odometer:          sql.NullFloat64{Float64: 12345.6, Valid: true},
			BatteryLevel:      sql.NullInt32{Int32: 85, Valid: true},
			UsableBatteryLevel: sql.NullInt32{Int32: 83, Valid: true},
			EstBatteryRange:   sql.NullFloat64{Float64: 380.2, Valid: true},
			RatedBatteryRange: sql.NullFloat64{Float64: 420.8, Valid: true},
			IdealBatteryRange: sql.NullFloat64{Float64: 400.5, Valid: true},
			OutsideTemp:       sql.NullFloat64{Float64: 18.5, Valid: true},
			InsideTemp:        sql.NullFloat64{Float64: 22.3, Valid: true},
			IsClimateOn:       sql.NullBool{Bool: true, Valid: true},
			IsPreconditioning: sql.NullBool{Bool: false, Valid: true},
			StateSince:        sql.NullTime{Time: now, Valid: true},
			IsCharging:        sql.NullBool{Bool: true, Valid: true},
			ChargingState:     sql.NullString{String: "charging", Valid: true},
			ChargerPower:      sql.NullInt32{Int32: 11000, Valid: true},
			ChargerVoltage:    sql.NullInt32{Int32: 240, Valid: true},
			ChargerPhases:     sql.NullInt32{Int32: 3, Valid: true},
			ChargerActualCurrent: sql.NullInt32{Int32: 45, Valid: true},
			ChargeEnergyAdded: sql.NullFloat64{Float64: 5.2, Valid: true},
			UnitOfLength:      sql.NullString{String: "km", Valid: true},
			UnitOfPressure:    sql.NullString{String: "bar", Valid: true},
			UnitOfTemperature: sql.NullString{String: "C", Valid: true},
		}

		response := mapper.MapToResponse(data, "online")

		// Verify basic car info
		if response.Data.Car.CarID != 1 {
			t.Errorf("Expected CarID 1, got %d", response.Data.Car.CarID)
		}

		if string(response.Data.Car.CarName) != "Test Car" {
			t.Errorf("Expected CarName 'Test Car', got %s", response.Data.Car.CarName)
		}

		// Verify status info
		if response.Data.Status.DisplayName != "Test Car" {
			t.Errorf("Expected DisplayName 'Test Car', got %s", response.Data.Status.DisplayName)
		}

		if response.Data.Status.State != "online" {
			t.Errorf("Expected State 'online', got %s", response.Data.Status.State)
		}

		if response.Data.Status.Odometer != 12345.6 {
			t.Errorf("Expected Odometer 12345.6, got %f", response.Data.Status.Odometer)
		}

		// Verify car details
		if response.Data.Status.CarDetails.Model != "Model 3" {
			t.Errorf("Expected Model 'Model 3', got %s", response.Data.Status.CarDetails.Model)
		}

		if response.Data.Status.CarDetails.TrimBadging != "Performance" {
			t.Errorf("Expected TrimBadging 'Performance', got %s", response.Data.Status.CarDetails.TrimBadging)
		}

		// Verify car exterior
		if response.Data.Status.CarExterior.ExteriorColor != "Red" {
			t.Errorf("Expected ExteriorColor 'Red', got %s", response.Data.Status.CarExterior.ExteriorColor)
		}

		// Verify location
		if response.Data.Status.CarGeodata.Location.Latitude != 37.7749 {
			t.Errorf("Expected Latitude 37.7749, got %f", response.Data.Status.CarGeodata.Location.Latitude)
		}

		if response.Data.Status.CarGeodata.Location.Longitude != -122.4194 {
			t.Errorf("Expected Longitude -122.4194, got %f", response.Data.Status.CarGeodata.Location.Longitude)
		}

		// Verify battery details
		if response.Data.Status.BatteryDetails.BatteryLevel != 85 {
			t.Errorf("Expected BatteryLevel 85, got %d", response.Data.Status.BatteryDetails.BatteryLevel)
		}

		if response.Data.Status.BatteryDetails.EstBatteryRange != 380.2 {
			t.Errorf("Expected EstBatteryRange 380.2, got %f", response.Data.Status.BatteryDetails.EstBatteryRange)
		}

		// Verify climate details
		if !response.Data.Status.ClimateDetails.IsClimateOn {
			t.Error("Expected IsClimateOn true, got false")
		}

		if response.Data.Status.ClimateDetails.OutsideTemp != 18.5 {
			t.Errorf("Expected OutsideTemp 18.5, got %f", response.Data.Status.ClimateDetails.OutsideTemp)
		}

		// Verify charging details - KEY FUNCTIONALITY
		if !response.Data.Status.ChargingDetails.PluggedIn {
			t.Error("Expected PluggedIn true, got false")
		}

		if response.Data.Status.ChargingDetails.ChargingState != "charging" {
			t.Errorf("Expected ChargingState 'charging', got %s", response.Data.Status.ChargingDetails.ChargingState)
		}

		if response.Data.Status.ChargingDetails.ChargerPower != 11000 {
			t.Errorf("Expected ChargerPower 11000, got %f", response.Data.Status.ChargingDetails.ChargerPower)
		}

		if response.Data.Status.ChargingDetails.ChargerVoltage != 240 {
			t.Errorf("Expected ChargerVoltage 240, got %d", response.Data.Status.ChargingDetails.ChargerVoltage)
		}

		if response.Data.Status.ChargingDetails.ChargeEnergyAdded != 5.2 {
			t.Errorf("Expected ChargeEnergyAdded 5.2, got %f", response.Data.Status.ChargingDetails.ChargeEnergyAdded)
		}

		// Verify units
		if response.Data.Units.UnitsLength != "km" {
			t.Errorf("Expected UnitsLength 'km', got %s", response.Data.Units.UnitsLength)
		}
	})

	t.Run("Null data handling", func(t *testing.T) {
		data := &CarStatusData{
			CarID: 2,
			// All other fields are null/invalid
		}

		response := mapper.MapToResponse(data, "unknown")

		// Verify defaults are applied
		if response.Data.Car.CarID != 2 {
			t.Errorf("Expected CarID 2, got %d", response.Data.Car.CarID)
		}

		if string(response.Data.Car.CarName) != "" {
			t.Errorf("Expected empty CarName, got %s", response.Data.Car.CarName)
		}

		if response.Data.Status.DisplayName != "Car 2" {
			t.Errorf("Expected DisplayName 'Car 2', got %s", response.Data.Status.DisplayName)
		}

		if response.Data.Status.State != "unknown" {
			t.Errorf("Expected State 'unknown', got %s", response.Data.Status.State)
		}

		// Verify charging defaults
		if response.Data.Status.ChargingDetails.PluggedIn {
			t.Error("Expected PluggedIn false for null data, got true")
		}

		if response.Data.Status.ChargingDetails.ChargingState != "disconnected" {
			t.Errorf("Expected ChargingState 'disconnected', got %s", response.Data.Status.ChargingDetails.ChargingState)
		}

		// Verify unit defaults
		if response.Data.Units.UnitsLength != "km" {
			t.Errorf("Expected default UnitsLength 'km', got %s", response.Data.Units.UnitsLength)
		}

		if response.Data.Units.UnitsPressure != "bar" {
			t.Errorf("Expected default UnitsPressure 'bar', got %s", response.Data.Units.UnitsPressure)
		}

		if response.Data.Units.UnitsTemperature != "C" {
			t.Errorf("Expected default UnitsTemperature 'C', got %s", response.Data.Units.UnitsTemperature)
		}
	})
}

func TestCarStatusMapper_ApplyUnitConversions(t *testing.T) {
	mapper := NewCarStatusMapper()

	t.Run("Miles conversion", func(t *testing.T) {
		response := &CarStatusResponse{}
		response.Data.Status.Odometer = 100.0
		response.Data.Status.BatteryDetails.EstBatteryRange = 400.0
		response.Data.Status.BatteryDetails.RatedBatteryRange = 420.0
		response.Data.Status.BatteryDetails.IdealBatteryRange = 410.0
		response.Data.Units.UnitsLength = "mi"

		mapper.ApplyUnitConversions(response)

		// Verify kilometers were converted to miles
		expectedOdometer := 100.0 * 0.62137119223733 // 62.137...
		if response.Data.Status.Odometer < expectedOdometer-0.01 || response.Data.Status.Odometer > expectedOdometer+0.01 {
			t.Errorf("Expected Odometer ~62.14, got %f", response.Data.Status.Odometer)
		}

		expectedRange := 400.0 * 0.62137119223733 // 248.548...
		if response.Data.Status.BatteryDetails.EstBatteryRange < expectedRange-0.01 || response.Data.Status.BatteryDetails.EstBatteryRange > expectedRange+0.01 {
			t.Errorf("Expected EstBatteryRange ~248.55, got %f", response.Data.Status.BatteryDetails.EstBatteryRange)
		}
	})

	t.Run("Fahrenheit conversion", func(t *testing.T) {
		response := &CarStatusResponse{}
		response.Data.Status.ClimateDetails.InsideTemp = 20.0   // 20C = 68F
		response.Data.Status.ClimateDetails.OutsideTemp = 0.0   // 0C = 32F
		response.Data.Units.UnitsTemperature = "F"

		mapper.ApplyUnitConversions(response)

		// Verify Celsius was converted to Fahrenheit
		if response.Data.Status.ClimateDetails.InsideTemp != 68.0 {
			t.Errorf("Expected InsideTemp 68.0F, got %f", response.Data.Status.ClimateDetails.InsideTemp)
		}

		if response.Data.Status.ClimateDetails.OutsideTemp != 32.0 {
			t.Errorf("Expected OutsideTemp 32.0F, got %f", response.Data.Status.ClimateDetails.OutsideTemp)
		}
	})

	t.Run("No conversion needed", func(t *testing.T) {
		response := &CarStatusResponse{}
		response.Data.Status.Odometer = 100.0
		response.Data.Status.ClimateDetails.InsideTemp = 20.0
		response.Data.Units.UnitsLength = "km"
		response.Data.Units.UnitsTemperature = "C"

		originalOdometer := response.Data.Status.Odometer
		originalTemp := response.Data.Status.ClimateDetails.InsideTemp

		mapper.ApplyUnitConversions(response)

		// Verify no changes were made
		if response.Data.Status.Odometer != originalOdometer {
			t.Errorf("Expected Odometer unchanged at %f, got %f", originalOdometer, response.Data.Status.Odometer)
		}

		if response.Data.Status.ClimateDetails.InsideTemp != originalTemp {
			t.Errorf("Expected InsideTemp unchanged at %f, got %f", originalTemp, response.Data.Status.ClimateDetails.InsideTemp)
		}
	})
}

func TestCarStatusMapper_HelperMethods(t *testing.T) {
	mapper := NewCarStatusMapper()

	t.Run("getDisplayName", func(t *testing.T) {
		// Test with name
		data1 := &CarStatusData{
			CarID: 1,
			Name:  sql.NullString{String: "My Tesla", Valid: true},
			Model: sql.NullString{String: "Model S", Valid: true},
		}
		displayName := mapper.getDisplayName(data1)
		if displayName != "My Tesla" {
			t.Errorf("Expected 'My Tesla', got %s", displayName)
		}

		// Test with model only
		data2 := &CarStatusData{
			CarID: 2,
			Name:  sql.NullString{Valid: false},
			Model: sql.NullString{String: "Model 3", Valid: true},
		}
		displayName = mapper.getDisplayName(data2)
		if displayName != "Model 3" {
			t.Errorf("Expected 'Model 3', got %s", displayName)
		}

		// Test with neither
		data3 := &CarStatusData{
			CarID: 3,
			Name:  sql.NullString{Valid: false},
			Model: sql.NullString{Valid: false},
		}
		displayName = mapper.getDisplayName(data3)
		if displayName != "Car 3" {
			t.Errorf("Expected 'Car 3', got %s", displayName)
		}
	})

	t.Run("getValue helpers", func(t *testing.T) {
		// Test string values
		validStr := sql.NullString{String: "test", Valid: true}
		invalidStr := sql.NullString{Valid: false}

		if mapper.getStringValue(validStr) != "test" {
			t.Errorf("Expected 'test', got %s", mapper.getStringValue(validStr))
		}

		if mapper.getStringValue(invalidStr) != "" {
			t.Errorf("Expected empty string, got %s", mapper.getStringValue(invalidStr))
		}

		// Test float values
		validFloat := sql.NullFloat64{Float64: 123.45, Valid: true}
		invalidFloat := sql.NullFloat64{Valid: false}

		if mapper.getFloat64Value(validFloat) != 123.45 {
			t.Errorf("Expected 123.45, got %f", mapper.getFloat64Value(validFloat))
		}

		if mapper.getFloat64Value(invalidFloat) != 0.0 {
			t.Errorf("Expected 0.0, got %f", mapper.getFloat64Value(invalidFloat))
		}

		// Test bool values
		validBool := sql.NullBool{Bool: true, Valid: true}
		invalidBool := sql.NullBool{Valid: false}

		if !mapper.getBoolValue(validBool) {
			t.Error("Expected true, got false")
		}

		if mapper.getBoolValue(invalidBool) {
			t.Error("Expected false, got true")
		}

		// Test int values
		validInt := sql.NullInt32{Int32: 42, Valid: true}
		invalidInt := sql.NullInt32{Valid: false}

		if mapper.getIntValue(validInt) != 42 {
			t.Errorf("Expected 42, got %d", mapper.getIntValue(validInt))
		}

		if mapper.getIntValue(invalidInt) != 0 {
			t.Errorf("Expected 0, got %d", mapper.getIntValue(invalidInt))
		}
	})
}