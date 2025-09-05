-- name: GetCarStatus :one
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
    cs.suspend_min,
    cs.suspend_after_idle_min,
    cs.req_not_unlocked,
    cs.free_supercharging,
    cs.use_streaming_api,
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
    p.locked,
    p.sentry_mode,
    p.windows_open,
    p.doors_open,
    p.trunk_open,
    p.frunk_open,
    p.is_user_present,
    p.shift_state,
    p.heading,
    p.elevation,
    -- Charging status
    COALESCE(p.charge_energy_added, 0) AS charge_energy_added,
    COALESCE(p.charger_power, 0) AS charger_power,
    COALESCE(p.charger_voltage, 0) AS charger_voltage,
    COALESCE(p.charger_phases, 0) AS charger_phases,
    COALESCE(p.charger_actual_current, 0) AS charger_actual_current,
    COALESCE(p.charge_current_request, 0) AS charge_current_request,
    COALESCE(p.charge_current_request_max, 0) AS charge_current_request_max,
    p.charge_limit_soc,
    p.charge_port_door_open,
    p.time_to_full_charge,
    p.scheduled_charging_start_time,
    CASE 
        WHEN cp.id IS NOT NULL AND cp.end_date IS NULL THEN 'charging'
        WHEN p.charge_port_door_open = true THEN 'plugged_in' 
        ELSE 'disconnected'
    END AS charging_state,
    CASE 
        WHEN cp.id IS NOT NULL AND cp.end_date IS NULL THEN true
        WHEN p.charge_port_door_open = true THEN true
        ELSE false
    END AS plugged_in,
    -- Latest state information
    s.state,
    s.start_date AS state_since,
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
WHERE c.id = $1;

-- name: GetCarBasicInfo :one
SELECT 
    id,
    name,
    vin,
    model
FROM cars 
WHERE id = $1;

-- name: IsCarConnectedToCharger :one
SELECT 
    CASE 
        WHEN cp.id IS NOT NULL AND cp.end_date IS NULL THEN true
        WHEN p.charge_port_door_open = true THEN true
        ELSE false
    END AS connected_to_charger,
    CASE 
        WHEN cp.id IS NOT NULL AND cp.end_date IS NULL THEN 'charging'
        WHEN p.charge_port_door_open = true THEN 'plugged_in' 
        ELSE 'disconnected'
    END AS connection_status,
    p.date AS last_update
FROM cars c
LEFT JOIN positions p ON c.id = p.car_id AND p.date = (
    SELECT MAX(date) FROM positions p2 WHERE p2.car_id = c.id
)
LEFT JOIN charging_processes cp ON c.id = cp.car_id AND cp.end_date IS NULL
WHERE c.id = $1;