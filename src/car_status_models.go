package main

import (
	"database/sql"
)

// CarStatusData holds raw database query results
type CarStatusData struct {
	// Car identification
	CarID       int             `db:"id"`
	Name        sql.NullString  `db:"name"`
	Model       sql.NullString  `db:"model"`
	TrimBadging sql.NullString  `db:"trim_badging"`
	ExteriorColor sql.NullString `db:"exterior_color"`
	WheelType   sql.NullString  `db:"wheel_type"`
	SpoilerType sql.NullString  `db:"spoiler_type"`
	Vin         sql.NullString  `db:"vin"`

	// Position and battery data
	PositionDate      sql.NullTime    `db:"position_date"`
	Latitude          sql.NullFloat64 `db:"latitude"`
	Longitude         sql.NullFloat64 `db:"longitude"`
	Speed             sql.NullInt32   `db:"speed"`
	Power             sql.NullInt32   `db:"power"`
	Odometer          sql.NullFloat64 `db:"odometer"`
	BatteryLevel      sql.NullInt32   `db:"battery_level"`
	UsableBatteryLevel sql.NullInt32  `db:"usable_battery_level"`
	IdealBatteryRange sql.NullFloat64 `db:"ideal_battery_range_km"`
	EstBatteryRange   sql.NullFloat64 `db:"est_battery_range_km"`
	RatedBatteryRange sql.NullFloat64 `db:"rated_battery_range_km"`

	// Climate data
	OutsideTemp       sql.NullFloat64 `db:"outside_temp"`
	InsideTemp        sql.NullFloat64 `db:"inside_temp"`
	IsClimateOn       sql.NullBool    `db:"is_climate_on"`
	IsPreconditioning sql.NullBool    `db:"is_preconditioning"`

	// State information
	State      sql.NullString `db:"state"`
	StateSince sql.NullTime   `db:"state_since"`

	// Charging information
	IsCharging           sql.NullBool    `db:"is_charging"`
	ChargingState        sql.NullString  `db:"charging_state"`
	ChargerPower         sql.NullInt32   `db:"charger_power"`
	ChargerVoltage       sql.NullInt32   `db:"charger_voltage"`
	ChargerPhases        sql.NullInt32   `db:"charger_phases"`
	ChargerActualCurrent sql.NullInt32   `db:"charger_actual_current"`
	ChargeEnergyAdded    sql.NullFloat64 `db:"charge_energy_added"`

	// Settings
	UnitOfLength      sql.NullString `db:"unit_of_length"`
	UnitOfPressure    sql.NullString `db:"unit_of_pressure"`
	UnitOfTemperature sql.NullString `db:"unit_of_temperature"`
}

// API Response structures (matching original API format)
type CarStatusResponse struct {
	Data CarStatusData_API `json:"data"`
}

type CarStatusData_API struct {
	Car    CarInfo_API    `json:"car"`
	Status StatusInfo_API `json:"status"`
	Units  UnitsInfo_API  `json:"units"`
}

type CarInfo_API struct {
	CarID   int        `json:"car_id"`
	CarName NullString `json:"car_name"`
}

type StatusInfo_API struct {
	DisplayName     string          `json:"display_name"`
	State           string          `json:"state"`
	StateSince      string          `json:"state_since"`
	Odometer        float64         `json:"odometer"`
	CarStatus       CarStatusInfo   `json:"car_status"`
	CarDetails      CarDetailsInfo  `json:"car_details"`
	CarExterior     CarExteriorInfo `json:"car_exterior"`
	CarGeodata      CarGeodataInfo  `json:"car_geodata"`
	ClimateDetails  ClimateInfo     `json:"climate_details"`
	BatteryDetails  BatteryInfo     `json:"battery_details"`
	ChargingDetails ChargingInfo    `json:"charging_details"`
}

type UnitsInfo_API struct {
	UnitsLength      string `json:"unit_of_length"`
	UnitsPressure    string `json:"unit_of_pressure"`
	UnitsTemperature string `json:"unit_of_temperature"`
}

type CarStatusInfo struct {
	Healthy       bool `json:"healthy"`
	Locked        bool `json:"locked"`
	SentryMode    bool `json:"sentry_mode"`
	WindowsOpen   bool `json:"windows_open"`
	DoorsOpen     bool `json:"doors_open"`
	TrunkOpen     bool `json:"trunk_open"`
	FrunkOpen     bool `json:"frunk_open"`
	IsUserPresent bool `json:"is_user_present"`
}

type CarDetailsInfo struct {
	Model       string `json:"model"`
	TrimBadging string `json:"trim_badging"`
}

type CarExteriorInfo struct {
	ExteriorColor string `json:"exterior_color"`
	SpoilerType   string `json:"spoiler_type"`
	WheelType     string `json:"wheel_type"`
}

type CarGeodataInfo struct {
	Location LocationInfo `json:"location"`
}

type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ClimateInfo struct {
	IsClimateOn       bool    `json:"is_climate_on"`
	InsideTemp        float64 `json:"inside_temp"`
	OutsideTemp       float64 `json:"outside_temp"`
	IsPreconditioning bool    `json:"is_preconditioning"`
}

type BatteryInfo struct {
	EstBatteryRange    float64 `json:"est_battery_range"`
	RatedBatteryRange  float64 `json:"rated_battery_range"`
	IdealBatteryRange  float64 `json:"ideal_battery_range"`
	BatteryLevel       int     `json:"battery_level"`
	UsableBatteryLevel int     `json:"usable_battery_level"`
}

type ChargingInfo struct {
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
}