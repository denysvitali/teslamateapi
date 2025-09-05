package main

import (
	"database/sql"
	"fmt"
	"time"
)

const carStatusQuery = `
	SELECT 
		c.id,
		c.name,
		c.model,
		c.trim_badging,
		c.exterior_color,
		c.wheel_type,
		c.spoiler_type,
		c.vin,
		-- Latest position data
		p.date AS position_date,
		p.latitude,
		p.longitude,
		p.speed,
		p.power,
		p.odometer,
		p.battery_level,
		p.usable_battery_level,
		p.ideal_battery_range_km,
		p.est_battery_range_km,
		p.rated_battery_range_km,
		p.outside_temp,
		p.inside_temp,
		p.is_climate_on,
		-- Latest state information
		s.state,
		s.start_date AS state_since,
		-- Check for active charging process
		CASE 
			WHEN cp.id IS NOT NULL AND cp.end_date IS NULL THEN true
			ELSE false
		END AS is_charging,
		CASE 
			WHEN cp.id IS NOT NULL AND cp.end_date IS NULL THEN 'charging'
			ELSE 'disconnected'
		END AS charging_state,
		-- Latest charge data if charging
		ch.charger_power,
		ch.charger_voltage,
		ch.charger_phases,
		ch.charger_actual_current,
		ch.charge_energy_added,
		-- Settings
		(SELECT unit_of_length FROM settings LIMIT 1) AS unit_of_length,
		(SELECT unit_of_pressure FROM settings LIMIT 1) AS unit_of_pressure,
		(SELECT unit_of_temperature FROM settings LIMIT 1) AS unit_of_temperature
	FROM cars c
	LEFT JOIN positions p ON c.id = p.car_id AND p.date = (
		SELECT MAX(date) FROM positions p2 WHERE p2.car_id = c.id
	)
	LEFT JOIN states s ON c.id = s.car_id AND s.start_date = (
		SELECT MAX(start_date) FROM states s2 WHERE s2.car_id = c.id
	)
	LEFT JOIN charging_processes cp ON c.id = cp.car_id AND cp.end_date IS NULL
	LEFT JOIN charges ch ON cp.id = ch.charging_process_id AND ch.date = (
		SELECT MAX(date) FROM charges ch2 WHERE ch2.charging_process_id = cp.id
	)
	WHERE c.id = $1`

// CarStatusService handles car status operations
type CarStatusService struct {
	db *sql.DB
}

func NewCarStatusService(database *sql.DB) *CarStatusService {
	return &CarStatusService{db: database}
}

// GetCarStatus retrieves comprehensive car status from database
func (s *CarStatusService) GetCarStatus(carID int) (*CarStatusData, error) {
	// Verify car exists first
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM cars WHERE id=$1)", carID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("car with ID %d does not exist", carID)
	}

	// Query comprehensive car status
	var data CarStatusData
	err = s.db.QueryRow(carStatusQuery, carID).Scan(
		&data.CarID,
		&data.Name,
		&data.Model,
		&data.TrimBadging,
		&data.ExteriorColor,
		&data.WheelType,
		&data.SpoilerType,
		&data.Vin,
		&data.PositionDate,
		&data.Latitude,
		&data.Longitude,
		&data.Speed,
		&data.Power,
		&data.Odometer,
		&data.BatteryLevel,
		&data.UsableBatteryLevel,
		&data.IdealBatteryRange,
		&data.EstBatteryRange,
		&data.RatedBatteryRange,
		&data.OutsideTemp,
		&data.InsideTemp,
		&data.IsClimateOn,
		&data.State,
		&data.StateSince,
		&data.IsCharging,
		&data.ChargingState,
		&data.ChargerPower,
		&data.ChargerVoltage,
		&data.ChargerPhases,
		&data.ChargerActualCurrent,
		&data.ChargeEnergyAdded,
		&data.UnitOfLength,
		&data.UnitOfPressure,
		&data.UnitOfTemperature,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no data available for car ID %d", carID)
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &data, nil
}

// DetermineVehicleState calculates vehicle state from available data
func (s *CarStatusService) DetermineVehicleState(data *CarStatusData) string {
	if data.State.Valid && data.State.String != "" {
		return data.State.String
	}

	// Fallback: determine state from position timestamp
	if data.PositionDate.Valid {
		timeDiff := time.Since(data.PositionDate.Time)
		switch {
		case timeDiff < 5*time.Minute:
			return "online"
		case timeDiff < 30*time.Minute:
			return "asleep"
		default:
			return "offline"
		}
	}

	return "unknown"
}