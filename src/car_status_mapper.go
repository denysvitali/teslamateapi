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

	// Car information
	response.Car.CarID = data.CarID
	response.Car.CarName = m.getStringValue(data.Name)

	// Status information
	response.Status.DisplayName = m.getDisplayName(data)
	response.Status.State = vehicleState
	response.Status.StateSince = m.getTimeValue(data.StateSince)
	response.Status.Odometer = m.getFloat64Value(data.Odometer)

	// Car details
	response.Status.CarDetails.Model = m.getStringValue(data.Model)
	response.Status.CarDetails.TrimBadging = m.getStringValue(data.TrimBadging)

	// Car exterior
	response.Status.CarExterior.ExteriorColor = m.getStringValue(data.ExteriorColor)
	response.Status.CarExterior.SpoilerType = m.getStringValue(data.SpoilerType)
	response.Status.CarExterior.WheelType = m.getStringValue(data.WheelType)

	// Geo data
	response.Status.CarGeodata.Latitude = m.getFloat64Value(data.Latitude)
	response.Status.CarGeodata.Longitude = m.getFloat64Value(data.Longitude)
	response.Status.CarGeodata.Geofence = "" // Not available in database

	// Physical status (not available from positions table, setting defaults)
	response.Status.CarStatus.Healthy = true
	response.Status.CarStatus.Locked = false
	response.Status.CarStatus.SentryMode = false
	response.Status.CarStatus.WindowsOpen = false
	response.Status.CarStatus.DoorsOpen = false
	response.Status.CarStatus.TrunkOpen = false
	response.Status.CarStatus.FrunkOpen = false
	response.Status.CarStatus.IsUserPresent = false

	// Car versions (not available from database)
	response.Status.CarVersions.UpdateAvailable = false
	response.Status.CarVersions.UpdateVersion = ""
	response.Status.CarVersions.Version = ""

	// Climate details
	response.Status.ClimateDetails.IsClimateOn = m.getBoolValue(data.IsClimateOn)
	response.Status.ClimateDetails.InsideTemp = m.getFloat64Value(data.InsideTemp)
	response.Status.ClimateDetails.OutsideTemp = m.getFloat64Value(data.OutsideTemp)
	response.Status.ClimateDetails.IsPreconditioning = false // Not available in database

	// Battery details
	response.Status.BatteryDetails.BatteryLevel = m.getIntValue(data.BatteryLevel)
	response.Status.BatteryDetails.EstBatteryRange = m.getFloat64Value(data.EstBatteryRange)
	response.Status.BatteryDetails.RatedBatteryRange = m.getFloat64Value(data.RatedBatteryRange)
	response.Status.BatteryDetails.IdealBatteryRange = m.getFloat64Value(data.IdealBatteryRange)
	response.Status.BatteryDetails.UsableBatteryLevel = m.getIntValue(data.UsableBatteryLevel)

	// Driving details
	response.Status.DrivingDetails.Elevation = m.getIntValue(data.Elevation)
	response.Status.DrivingDetails.Heading = 0 // Not available in database
	response.Status.DrivingDetails.Power = m.getIntValue(data.Power)
	response.Status.DrivingDetails.ShiftState = "" // Not available in database
	response.Status.DrivingDetails.Speed = m.getIntValue(data.Speed)

	// Charging details
	response.Status.ChargingDetails.PluggedIn = m.getBoolValue(data.IsCharging)
	response.Status.ChargingDetails.ChargeEnergyAdded = float32(m.getFloat64Value(data.ChargeEnergyAdded))
	response.Status.ChargingDetails.ChargerActualCurrent = float32(m.getIntValue(data.ChargerActualCurrent))
	response.Status.ChargingDetails.ChargerPhases = m.getIntValue(data.ChargerPhases)
	response.Status.ChargingDetails.ChargerPower = float32(m.getIntValue(data.ChargerPower))
	response.Status.ChargingDetails.ChargerVoltage = float32(m.getIntValue(data.ChargerVoltage))
	response.Status.ChargingDetails.ChargePortDoorOpen = m.getBoolValue(data.IsCharging) // Approximate

	// Fields not available in database - set to defaults
	response.Status.ChargingDetails.ChargeCurrentRequest = 0
	response.Status.ChargingDetails.ChargeCurrentRequestMax = 0
	response.Status.ChargingDetails.ChargeLimitSoc = 0
	response.Status.ChargingDetails.ScheduledChargingStartTime = time.Time{}
	response.Status.ChargingDetails.TimeToFullCharge = 0

	// TPMS details
	response.Status.TpmsDetails.TpmsPressureFl = m.getFloat64Value(data.TpmsPressureFl)
	response.Status.TpmsDetails.TpmsPressureFr = m.getFloat64Value(data.TpmsPressureFr)
	response.Status.TpmsDetails.TpmsPressureRl = m.getFloat64Value(data.TpmsPressureRl)
	response.Status.TpmsDetails.TpmsPressureRr = m.getFloat64Value(data.TpmsPressureRr)

	// Units
	response.Units.UnitOfLength = m.getStringValueWithDefault(data.UnitOfLength, "km")
	response.Units.UnitOfTemperature = m.getStringValueWithDefault(data.UnitOfTemperature, "C")

	return response
}

// ApplyUnitConversions applies unit conversions based on user preferences
func (m *CarStatusMapper) ApplyUnitConversions(response *CarStatusResponse) {
	// Length conversions
	if response.Units.UnitOfLength == "mi" {
		response.Status.Odometer = kilometersToMiles(response.Status.Odometer)
		response.Status.BatteryDetails.EstBatteryRange = kilometersToMiles(response.Status.BatteryDetails.EstBatteryRange)
		response.Status.BatteryDetails.RatedBatteryRange = kilometersToMiles(response.Status.BatteryDetails.RatedBatteryRange)
		response.Status.BatteryDetails.IdealBatteryRange = kilometersToMiles(response.Status.BatteryDetails.IdealBatteryRange)
	}

	// Temperature conversions
	if response.Units.UnitOfTemperature == "F" {
		response.Status.ClimateDetails.InsideTemp = celsiusToFahrenheit(response.Status.ClimateDetails.InsideTemp)
		response.Status.ClimateDetails.OutsideTemp = celsiusToFahrenheit(response.Status.ClimateDetails.OutsideTemp)
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

func (m *CarStatusMapper) getTimeValue(nullTime sql.NullTime) time.Time {
	if nullTime.Valid {
		return nullTime.Time
	}
	return time.Now()
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