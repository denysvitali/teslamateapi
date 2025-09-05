package main

import (
	"database/sql"
	"fmt"
	"time"
)

// CarStatusMapper converts database results to API response format
type CarStatusMapper struct{}

func NewCarStatusMapper() *CarStatusMapper {
	return &CarStatusMapper{}
}

// MapToResponse converts CarStatusData to CarStatusResponse
func (m *CarStatusMapper) MapToResponse(data *CarStatusData, vehicleState string) *CarStatusResponse {
	response := &CarStatusResponse{}

	// Car identification
	response.Data.Car.CarID = data.CarID
	if data.Name.Valid {
		response.Data.Car.CarName = NullString(data.Name.String)
	} else {
		response.Data.Car.CarName = NullString("")
	}

	// Display name
	response.Data.Status.DisplayName = m.getDisplayName(data)

	// State information
	response.Data.Status.State = vehicleState
	response.Data.Status.StateSince = m.getFormattedTime(data.StateSince)

	// Basic car info
	response.Data.Status.Odometer = m.getFloat64Value(data.Odometer)

	// Car status (limited data from positions table)
	response.Data.Status.CarStatus = CarStatusInfo{
		Healthy:       true, // Assume healthy if we have data
		Locked:        false, // Not available in positions table
		SentryMode:    false, // Not available
		WindowsOpen:   false, // Not available
		DoorsOpen:     false, // Not available
		TrunkOpen:     false, // Not available
		FrunkOpen:     false, // Not available
		IsUserPresent: false, // Not available
	}

	// Car details
	response.Data.Status.CarDetails = CarDetailsInfo{
		Model:       m.getStringValue(data.Model),
		TrimBadging: m.getStringValue(data.TrimBadging),
	}

	// Car exterior
	response.Data.Status.CarExterior = CarExteriorInfo{
		ExteriorColor: m.getStringValue(data.ExteriorColor),
		SpoilerType:   m.getStringValue(data.SpoilerType),
		WheelType:     m.getStringValue(data.WheelType),
	}

	// Location
	response.Data.Status.CarGeodata.Location.Latitude = m.getFloat64Value(data.Latitude)
	response.Data.Status.CarGeodata.Location.Longitude = m.getFloat64Value(data.Longitude)

	// Climate details
	response.Data.Status.ClimateDetails = ClimateInfo{
		IsClimateOn: m.getBoolValue(data.IsClimateOn),
		InsideTemp:  m.getFloat64Value(data.InsideTemp),
		OutsideTemp: m.getFloat64Value(data.OutsideTemp),
	}

	// Battery details
	response.Data.Status.BatteryDetails = BatteryInfo{
		EstBatteryRange:    m.getFloat64Value(data.EstBatteryRange),
		RatedBatteryRange:  m.getFloat64Value(data.RatedBatteryRange),
		IdealBatteryRange:  m.getFloat64Value(data.IdealBatteryRange),
		BatteryLevel:       m.getIntValue(data.BatteryLevel),
		UsableBatteryLevel: m.getIntValue(data.UsableBatteryLevel),
	}

	// Charging details - KEY FEATURE for user requirements
	response.Data.Status.ChargingDetails = ChargingInfo{
		PluggedIn:                  m.getBoolValue(data.IsCharging),
		ChargingState:              m.getChargingState(data.ChargingState),
		ChargeEnergyAdded:          m.getFloat64Value(data.ChargeEnergyAdded),
		ChargeLimitSoc:             0, // Not available without Tesla API
		ChargePortDoorOpen:         m.getBoolValue(data.IsCharging), // Approximate
		ChargerActualCurrent:       float64(m.getIntValue(data.ChargerActualCurrent)),
		ChargerPhases:              m.getIntValue(data.ChargerPhases),
		ChargerPower:               float64(m.getIntValue(data.ChargerPower)),
		ChargerVoltage:             m.getIntValue(data.ChargerVoltage),
		ChargeCurrentRequest:       0, // Not available without Tesla API
		ChargeCurrentRequestMax:    0, // Not available without Tesla API
		ScheduledChargingStartTime: "", // Not available without Tesla API
		TimeToFullCharge:           0.0, // Not available without Tesla API
	}

	// Units
	response.Data.Units.UnitsLength = m.getStringValueWithDefault(data.UnitOfLength, "km")
	response.Data.Units.UnitsPressure = m.getStringValueWithDefault(data.UnitOfPressure, "bar")
	response.Data.Units.UnitsTemperature = m.getStringValueWithDefault(data.UnitOfTemperature, "C")

	return response
}

// ApplyUnitConversions applies unit conversions based on user preferences
func (m *CarStatusMapper) ApplyUnitConversions(response *CarStatusResponse) {
	// Length conversions
	if response.Data.Units.UnitsLength == "mi" {
		response.Data.Status.Odometer = kilometersToMiles(response.Data.Status.Odometer)
		response.Data.Status.BatteryDetails.EstBatteryRange = kilometersToMiles(response.Data.Status.BatteryDetails.EstBatteryRange)
		response.Data.Status.BatteryDetails.RatedBatteryRange = kilometersToMiles(response.Data.Status.BatteryDetails.RatedBatteryRange)
		response.Data.Status.BatteryDetails.IdealBatteryRange = kilometersToMiles(response.Data.Status.BatteryDetails.IdealBatteryRange)
	}

	// Temperature conversions
	if response.Data.Units.UnitsTemperature == "F" {
		response.Data.Status.ClimateDetails.InsideTemp = celsiusToFahrenheit(response.Data.Status.ClimateDetails.InsideTemp)
		response.Data.Status.ClimateDetails.OutsideTemp = celsiusToFahrenheit(response.Data.Status.ClimateDetails.OutsideTemp)
	}
}

// Helper methods for safe value extraction
func (m *CarStatusMapper) getDisplayName(data *CarStatusData) string {
	if data.Name.Valid && data.Name.String != "" {
		return data.Name.String
	}
	if data.Model.Valid {
		return data.Model.String
	}
	return fmt.Sprintf("Car %d", data.CarID)
}

func (m *CarStatusMapper) getFormattedTime(nullTime sql.NullTime) string {
	if nullTime.Valid {
		return getTimeInTimeZone(nullTime.Time.Format(dbTimestampFormat))
	}
	return getTimeInTimeZone(time.Now().Format(dbTimestampFormat))
}

func (m *CarStatusMapper) getStringValue(nullStr sql.NullString) string {
	if nullStr.Valid {
		return nullStr.String
	}
	return ""
}

func (m *CarStatusMapper) getStringValueWithDefault(nullStr sql.NullString, defaultVal string) string {
	if nullStr.Valid {
		return nullStr.String
	}
	return defaultVal
}

func (m *CarStatusMapper) getFloat64Value(nullFloat sql.NullFloat64) float64 {
	if nullFloat.Valid {
		return nullFloat.Float64
	}
	return 0.0
}

func (m *CarStatusMapper) getIntValue(nullInt sql.NullInt32) int {
	if nullInt.Valid {
		return int(nullInt.Int32)
	}
	return 0
}

func (m *CarStatusMapper) getBoolValue(nullBool sql.NullBool) bool {
	if nullBool.Valid {
		return nullBool.Bool
	}
	return false
}

func (m *CarStatusMapper) getChargingState(nullStr sql.NullString) string {
	if nullStr.Valid && nullStr.String != "" {
		return nullStr.String
	}
	return "disconnected"
}