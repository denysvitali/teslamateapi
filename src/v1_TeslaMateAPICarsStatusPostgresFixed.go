package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// TeslaMateAPICarsStatusPostgresV1Fixed provides car status directly from PostgreSQL (corrected version)
func TeslaMateAPICarsStatusPostgresV1Fixed(c *gin.Context) {
	// Get CarID param from URL
	carID := convertStringToInteger(c.Param("CarID"))

	// First, verify the car exists
	var carExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM cars WHERE id=$1)", carID).Scan(&carExists)
	if err != nil {
		TeslaMateAPIHandleErrorResponse(c, "TeslaMateAPICarsStatusPostgresV1Fixed", "Database error", err.Error())
		return
	}

	if !carExists {
		TeslaMateAPIHandleErrorResponse(c, "TeslaMateAPICarsStatusPostgresV1Fixed", "Car not found", fmt.Sprintf("Car with ID %d does not exist", carID))
		return
	}

	// Query for car basic info and latest position
	query := `
		SELECT 
			c.id,
			c.name,
			c.model,
			c.trim_badging,
			c.exterior_color,
			c.wheel_type,
			c.spoiler_type,
			c.vin,
			c.eid,
			c.vid,
			COALESCE(cs.suspend_min, 21) as suspend_min,
			COALESCE(cs.suspend_after_idle_min, 15) as suspend_after_idle_min,
			COALESCE(cs.req_not_unlocked, false) as req_not_unlocked,
			COALESCE(cs.free_supercharging, false) as free_supercharging,
			COALESCE(cs.use_streaming_api, false) as use_streaming_api,
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
			p.is_preconditioning,
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
		LEFT JOIN car_settings cs ON c.id = cs.id
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
		WHERE c.id = $1;`

	type StatusResult struct {
		CarID                   int             `db:"id"`
		Name                    sql.NullString  `db:"name"`
		Model                   sql.NullString  `db:"model"`
		TrimBadging            sql.NullString  `db:"trim_badging"`
		ExteriorColor          sql.NullString  `db:"exterior_color"`
		WheelType              sql.NullString  `db:"wheel_type"`
		SpoilerType            sql.NullString  `db:"spoiler_type"`
		Vin                    sql.NullString  `db:"vin"`
		EID                    sql.NullInt64   `db:"eid"`
		VID                    sql.NullInt64   `db:"vid"`
		SuspendMin             sql.NullInt32   `db:"suspend_min"`
		SuspendAfterIdleMin    sql.NullInt32   `db:"suspend_after_idle_min"`
		ReqNotUnlocked         sql.NullBool    `db:"req_not_unlocked"`
		FreeSupercharging      sql.NullBool    `db:"free_supercharging"`
		UseStreamingAPI        sql.NullBool    `db:"use_streaming_api"`
		PositionDate           sql.NullTime    `db:"position_date"`
		Latitude               sql.NullFloat64 `db:"latitude"`
		Longitude              sql.NullFloat64 `db:"longitude"`
		Speed                  sql.NullInt32   `db:"speed"`
		Power                  sql.NullInt32   `db:"power"`
		Odometer               sql.NullFloat64 `db:"odometer"`
		BatteryLevel           sql.NullInt32   `db:"battery_level"`
		UsableBatteryLevel     sql.NullInt32   `db:"usable_battery_level"`
		IdealBatteryRange      sql.NullFloat64 `db:"ideal_battery_range_km"`
		EstBatteryRange        sql.NullFloat64 `db:"est_battery_range_km"`
		RatedBatteryRange      sql.NullFloat64 `db:"rated_battery_range_km"`
		OutsideTemp            sql.NullFloat64 `db:"outside_temp"`
		InsideTemp             sql.NullFloat64 `db:"inside_temp"`
		IsClimateOn            sql.NullBool    `db:"is_climate_on"`
		IsPreconditioning      sql.NullBool    `db:"is_preconditioning"`
		State                  sql.NullString  `db:"state"`
		StateSince             sql.NullTime    `db:"state_since"`
		IsCharging             sql.NullBool    `db:"is_charging"`
		ChargingState          sql.NullString  `db:"charging_state"`
		ChargerPower           sql.NullInt32   `db:"charger_power"`
		ChargerVoltage         sql.NullInt32   `db:"charger_voltage"`
		ChargerPhases          sql.NullInt32   `db:"charger_phases"`
		ChargerActualCurrent   sql.NullInt32   `db:"charger_actual_current"`
		ChargeEnergyAdded      sql.NullFloat64 `db:"charge_energy_added"`
		UnitOfLength           sql.NullString  `db:"unit_of_length"`
		UnitOfPressure         sql.NullString  `db:"unit_of_pressure"`
		UnitOfTemperature      sql.NullString  `db:"unit_of_temperature"`
	}

	var result StatusResult
	err = db.QueryRow(query, carID).Scan(
		&result.CarID,
		&result.Name,
		&result.Model,
		&result.TrimBadging,
		&result.ExteriorColor,
		&result.WheelType,
		&result.SpoilerType,
		&result.Vin,
		&result.EID,
		&result.VID,
		&result.SuspendMin,
		&result.SuspendAfterIdleMin,
		&result.ReqNotUnlocked,
		&result.FreeSupercharging,
		&result.UseStreamingAPI,
		&result.PositionDate,
		&result.Latitude,
		&result.Longitude,
		&result.Speed,
		&result.Power,
		&result.Odometer,
		&result.BatteryLevel,
		&result.UsableBatteryLevel,
		&result.IdealBatteryRange,
		&result.EstBatteryRange,
		&result.RatedBatteryRange,
		&result.OutsideTemp,
		&result.InsideTemp,
		&result.IsClimateOn,
		&result.IsPreconditioning,
		&result.State,
		&result.StateSince,
		&result.IsCharging,
		&result.ChargingState,
		&result.ChargerPower,
		&result.ChargerVoltage,
		&result.ChargerPhases,
		&result.ChargerActualCurrent,
		&result.ChargeEnergyAdded,
		&result.UnitOfLength,
		&result.UnitOfPressure,
		&result.UnitOfTemperature,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			TeslaMateAPIHandleErrorResponse(c, "TeslaMateAPICarsStatusPostgresV1Fixed", "Car not found in database", fmt.Sprintf("No data available for car ID %d", carID))
			return
		}
		TeslaMateAPIHandleErrorResponse(c, "TeslaMateAPICarsStatusPostgresV1Fixed", "Database query failed", err.Error())
		return
	}

	// Create response structure that matches the original API format
	type StatusResponse struct {
		Data struct {
			Car struct {
				CarID   int        `json:"car_id"`
				CarName NullString `json:"car_name"`
			} `json:"car"`
			Status struct {
				DisplayName     string `json:"display_name"`
				State           string `json:"state"`
				StateSince      string `json:"state_since"`
				Odometer        float64 `json:"odometer"`
				CarStatus       struct {
					Healthy                bool `json:"healthy"`
					Locked                 bool `json:"locked"`
					SentryMode             bool `json:"sentry_mode"`
					WindowsOpen            bool `json:"windows_open"`
					DoorsOpen              bool `json:"doors_open"`
					TrunkOpen              bool `json:"trunk_open"`
					FrunkOpen              bool `json:"frunk_open"`
					IsUserPresent          bool `json:"is_user_present"`
				} `json:"car_status"`
				CarDetails      struct {
					Model       string `json:"model"`
					TrimBadging string `json:"trim_badging"`
				} `json:"car_details"`
				CarExterior     struct {
					ExteriorColor string `json:"exterior_color"`
					SpoilerType   string `json:"spoiler_type"`
					WheelType     string `json:"wheel_type"`
				} `json:"car_exterior"`
				CarGeodata      struct {
					Location struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
					} `json:"location"`
				} `json:"car_geodata"`
				DrivingDetails  struct {
					ShiftState string `json:"shift_state"`
					Power      int    `json:"power"`
					Speed      int    `json:"speed"`
					Heading    int    `json:"heading"`
					Elevation  int    `json:"elevation"`
				} `json:"driving_details"`
				ClimateDetails  struct {
					IsClimateOn       bool    `json:"is_climate_on"`
					InsideTemp        float64 `json:"inside_temp"`
					OutsideTemp       float64 `json:"outside_temp"`
					IsPreconditioning bool    `json:"is_preconditioning"`
				} `json:"climate_details"`
				BatteryDetails  struct {
					EstBatteryRange    float64 `json:"est_battery_range"`
					RatedBatteryRange  float64 `json:"rated_battery_range"`
					IdealBatteryRange  float64 `json:"ideal_battery_range"`
					BatteryLevel       int     `json:"battery_level"`
					UsableBatteryLevel int     `json:"usable_battery_level"`
				} `json:"battery_details"`
				ChargingDetails struct {
					PluggedIn                  bool    `json:"plugged_in"`
					ChargingState              string  `json:"charging_state"`
					ChargeEnergyAdded          float64 `json:"charge_energy_added"`
					ChargeLimitSoc             int     `json:"charge_limit_soc"`
					ChargePortDoorOpen         bool    `json:"charge_port_door_open"`
					ChargerActualCurrent       float64 `json:"charger_actual_current"`
					ChargerPhases              int     `json:"charger_phases"`
					ChargerPower               float64 `json:"charger_power"`
					ChargerVoltage             int     `json:"charger_voltage"`
					ChargeCurrentRequest       int     `json:"charge_current_request"`
					ChargeCurrentRequestMax    int     `json:"charge_current_request_max"`
					ScheduledChargingStartTime string  `json:"scheduled_charging_start_time"`
					TimeToFullCharge           float64 `json:"time_to_full_charge"`
				} `json:"charging_details"`
			} `json:"status"`
			Units struct {
				UnitsLength      string `json:"unit_of_length"`
				UnitsPressure    string `json:"unit_of_pressure"`
				UnitsTemperature string `json:"unit_of_temperature"`
			} `json:"units"`
		} `json:"data"`
	}

	// Build response
	response := StatusResponse{}
	response.Data.Car.CarID = result.CarID
	if result.Name.Valid {
		response.Data.Car.CarName = NullString(result.Name.String)
	} else {
		response.Data.Car.CarName = NullString("")
	}
	
	// Set display name (use name if available, otherwise model)
	if result.Name.Valid && result.Name.String != "" {
		response.Data.Status.DisplayName = result.Name.String
	} else if result.Model.Valid {
		response.Data.Status.DisplayName = result.Model.String
	} else {
		response.Data.Status.DisplayName = fmt.Sprintf("Car %d", carID)
	}
	
	// State information - this addresses the main issue
	if result.State.Valid {
		response.Data.Status.State = result.State.String
	} else {
		// Determine state based on available data
		if result.PositionDate.Valid {
			timeDiff := time.Since(result.PositionDate.Time)
			if timeDiff < 5*time.Minute {
				response.Data.Status.State = "online"
			} else if timeDiff < 30*time.Minute {
				response.Data.Status.State = "asleep"
			} else {
				response.Data.Status.State = "offline"
			}
		} else {
			response.Data.Status.State = "unknown"
		}
	}
	
	if result.StateSince.Valid {
		response.Data.Status.StateSince = getTimeInTimeZone(result.StateSince.Time.Format(dbTimestampFormat))
	} else {
		response.Data.Status.StateSince = getTimeInTimeZone(time.Now().Format(dbTimestampFormat))
	}
	
	// Odometer
	if result.Odometer.Valid {
		response.Data.Status.Odometer = result.Odometer.Float64
	}
	
	// Car status - assume healthy since we have database data
	response.Data.Status.CarStatus.Healthy = true
	response.Data.Status.CarStatus.Locked = false  // Not available in positions table
	response.Data.Status.CarStatus.SentryMode = false // Not available in positions table
	response.Data.Status.CarStatus.WindowsOpen = false // Not available in positions table
	response.Data.Status.CarStatus.DoorsOpen = false // Not available in positions table
	response.Data.Status.CarStatus.TrunkOpen = false // Not available in positions table
	response.Data.Status.CarStatus.FrunkOpen = false // Not available in positions table
	response.Data.Status.CarStatus.IsUserPresent = false // Not available in positions table
	
	// Car details
	if result.Model.Valid {
		response.Data.Status.CarDetails.Model = result.Model.String
	}
	if result.TrimBadging.Valid {
		response.Data.Status.CarDetails.TrimBadging = result.TrimBadging.String
	}
	
	// Car exterior
	if result.ExteriorColor.Valid {
		response.Data.Status.CarExterior.ExteriorColor = result.ExteriorColor.String
	}
	if result.SpoilerType.Valid {
		response.Data.Status.CarExterior.SpoilerType = result.SpoilerType.String
	}
	if result.WheelType.Valid {
		response.Data.Status.CarExterior.WheelType = result.WheelType.String
	}
	
	// Location
	if result.Latitude.Valid {
		response.Data.Status.CarGeodata.Location.Latitude = result.Latitude.Float64
	}
	if result.Longitude.Valid {
		response.Data.Status.CarGeodata.Location.Longitude = result.Longitude.Float64
	}
	
	// Driving details
	response.Data.Status.DrivingDetails.ShiftState = "" // Not available in positions table
	if result.Power.Valid {
		response.Data.Status.DrivingDetails.Power = int(result.Power.Int32)
	}
	if result.Speed.Valid {
		response.Data.Status.DrivingDetails.Speed = int(result.Speed.Int32)
	}
	response.Data.Status.DrivingDetails.Heading = 0 // Not available in current positions table
	response.Data.Status.DrivingDetails.Elevation = 0 // Called altitude in positions table
	
	// Climate details
	response.Data.Status.ClimateDetails.IsClimateOn = result.IsClimateOn.Valid && result.IsClimateOn.Bool
	if result.InsideTemp.Valid {
		response.Data.Status.ClimateDetails.InsideTemp = result.InsideTemp.Float64
	}
	if result.OutsideTemp.Valid {
		response.Data.Status.ClimateDetails.OutsideTemp = result.OutsideTemp.Float64
	}
	response.Data.Status.ClimateDetails.IsPreconditioning = result.IsPreconditioning.Valid && result.IsPreconditioning.Bool
	
	// Battery details
	if result.EstBatteryRange.Valid {
		response.Data.Status.BatteryDetails.EstBatteryRange = result.EstBatteryRange.Float64
	}
	if result.RatedBatteryRange.Valid {
		response.Data.Status.BatteryDetails.RatedBatteryRange = result.RatedBatteryRange.Float64
	}
	if result.IdealBatteryRange.Valid {
		response.Data.Status.BatteryDetails.IdealBatteryRange = result.IdealBatteryRange.Float64
	}
	if result.BatteryLevel.Valid {
		response.Data.Status.BatteryDetails.BatteryLevel = int(result.BatteryLevel.Int32)
	}
	if result.UsableBatteryLevel.Valid {
		response.Data.Status.BatteryDetails.UsableBatteryLevel = int(result.UsableBatteryLevel.Int32)
	}
	
	// Charging details - this is the KEY FEATURE for the user's requirement
	response.Data.Status.ChargingDetails.PluggedIn = result.IsCharging.Valid && result.IsCharging.Bool
	if result.ChargingState.Valid {
		response.Data.Status.ChargingDetails.ChargingState = result.ChargingState.String
	} else {
		response.Data.Status.ChargingDetails.ChargingState = "disconnected"
	}
	if result.ChargeEnergyAdded.Valid {
		response.Data.Status.ChargingDetails.ChargeEnergyAdded = result.ChargeEnergyAdded.Float64
	}
	response.Data.Status.ChargingDetails.ChargeLimitSoc = 0 // Not available without Tesla API data
	response.Data.Status.ChargingDetails.ChargePortDoorOpen = result.IsCharging.Valid && result.IsCharging.Bool
	if result.ChargerActualCurrent.Valid {
		response.Data.Status.ChargingDetails.ChargerActualCurrent = float64(result.ChargerActualCurrent.Int32)
	}
	if result.ChargerPhases.Valid {
		response.Data.Status.ChargingDetails.ChargerPhases = int(result.ChargerPhases.Int32)
	}
	if result.ChargerPower.Valid {
		response.Data.Status.ChargingDetails.ChargerPower = float64(result.ChargerPower.Int32)
	}
	if result.ChargerVoltage.Valid {
		response.Data.Status.ChargingDetails.ChargerVoltage = int(result.ChargerVoltage.Int32)
	}
	response.Data.Status.ChargingDetails.ChargeCurrentRequest = 0 // Not available without Tesla API data
	response.Data.Status.ChargingDetails.ChargeCurrentRequestMax = 0 // Not available without Tesla API data
	response.Data.Status.ChargingDetails.ScheduledChargingStartTime = "" // Not available without Tesla API data
	response.Data.Status.ChargingDetails.TimeToFullCharge = 0.0 // Not available without Tesla API data
	
	// Units
	if result.UnitOfLength.Valid {
		response.Data.Units.UnitsLength = result.UnitOfLength.String
	} else {
		response.Data.Units.UnitsLength = "km"
	}
	if result.UnitOfPressure.Valid {
		response.Data.Units.UnitsPressure = result.UnitOfPressure.String
	} else {
		response.Data.Units.UnitsPressure = "bar"
	}
	if result.UnitOfTemperature.Valid {
		response.Data.Units.UnitsTemperature = result.UnitOfTemperature.String
	} else {
		response.Data.Units.UnitsTemperature = "C"
	}

	// Apply unit conversions if needed
	if response.Data.Units.UnitsLength == "mi" {
		response.Data.Status.Odometer = kilometersToMiles(response.Data.Status.Odometer)
		response.Data.Status.BatteryDetails.EstBatteryRange = kilometersToMiles(response.Data.Status.BatteryDetails.EstBatteryRange)
		response.Data.Status.BatteryDetails.RatedBatteryRange = kilometersToMiles(response.Data.Status.BatteryDetails.RatedBatteryRange)
		response.Data.Status.BatteryDetails.IdealBatteryRange = kilometersToMiles(response.Data.Status.BatteryDetails.IdealBatteryRange)
		response.Data.Status.DrivingDetails.Speed = kilometersToMilesInteger(response.Data.Status.DrivingDetails.Speed)
	}
	if response.Data.Units.UnitsTemperature == "F" {
		response.Data.Status.ClimateDetails.InsideTemp = celsiusToFahrenheit(response.Data.Status.ClimateDetails.InsideTemp)
		response.Data.Status.ClimateDetails.OutsideTemp = celsiusToFahrenheit(response.Data.Status.ClimateDetails.OutsideTemp)
	}

	if gin.IsDebugging() {
		log.Printf("[debug] TeslaMateAPICarsStatusPostgresV1Fixed - Car %d: charging_state=%s, plugged_in=%t, is_charging_from_db=%t", 
			carID, response.Data.Status.ChargingDetails.ChargingState, response.Data.Status.ChargingDetails.PluggedIn, 
			result.IsCharging.Valid && result.IsCharging.Bool)
	}

	TeslaMateAPIHandleSuccessResponse(c, "TeslaMateAPICarsStatusPostgresV1Fixed", response)
}