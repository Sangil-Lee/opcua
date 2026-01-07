-- ===================================================================
-- PLC Logic Script for Virtual PLC Simulator
-- ===================================================================
-- This script implements PLC control logic using sensor/actuator tags
-- Generated automatically from sensors.json
--
-- Scan Time: 100ms (configurable)
-- Execution: run_logic() is called every scan cycle
-- ===================================================================

-- ===================================================================
-- Global Variables (persist across scans)
-- ===================================================================
scan_count = scan_count or 0
system_start_time = system_start_time or 0
alarm_count = alarm_count or 0

-- Temperature control thresholds
TEMP_LOW_THRESHOLD = 20.0
TEMP_HIGH_THRESHOLD = 80.0
TEMP_SETPOINT = 50.0

-- Pressure control thresholds
PRESSURE_HIGH_LIMIT = 9.0

-- Vibration alarm threshold
VIBRATION_ALARM_LEVEL = 12.0

-- ===================================================================
-- Initialization Function (called once at startup)
-- ===================================================================
function init()
    plc_log("========================================")
    plc_log("PLC Logic Initialization Started")
    plc_log("========================================")

    system_start_time = get_time()

    -- Count available tags
    local tag_count = 0
    for k, v in pairs(Data) do
        tag_count = tag_count + 1
    end

    plc_log(string.format("Total Tags Available: %d", tag_count))
    plc_log("System initialized successfully")
    plc_log("========================================")
end

-- ===================================================================
-- Main Logic Function (called every scan cycle)
-- ===================================================================
function run_logic()
    scan_count = scan_count + 1

    -- Log system status every 1000 scans (~100 seconds at 100ms scan)
    if scan_count % 1000 == 0 then
        local runtime = get_time() - system_start_time
        plc_log(string.format("=== PLC Status: Scan #%d, Runtime: %ds, Alarms: %d ===",
                              scan_count, runtime, alarm_count))
    end

    -- ===================================================================
    -- Temperature Control Logic
    -- ===================================================================
    if Data.TemperatureSensor_Tank1 and Data.RelayActuator_Heater then
        local temp = Data.TemperatureSensor_Tank1

        -- Simple ON/OFF control with hysteresis
        if temp < (TEMP_SETPOINT - 5.0) then
            Data.RelayActuator_Heater = true  -- Turn ON heater
        elseif temp > (TEMP_SETPOINT + 5.0) then
            Data.RelayActuator_Heater = false -- Turn OFF heater
        end
        -- Between setpoint ± 5°C: maintain current state

        -- Temperature alarms
        if temp < TEMP_LOW_THRESHOLD then
            if scan_count % 500 == 0 then
                plc_log(string.format("ALARM: Tank1 Temperature LOW: %.2f°C", temp))
                alarm_count = alarm_count + 1
            end
        elseif temp > TEMP_HIGH_THRESHOLD then
            if scan_count % 500 == 0 then
                plc_log(string.format("ALARM: Tank1 Temperature HIGH: %.2f°C", temp))
                alarm_count = alarm_count + 1
            end
            -- Emergency: force heater OFF
            Data.RelayActuator_Heater = false
        end
    end

    -- ===================================================================
    -- Pressure Control Logic
    -- ===================================================================
    if Data.PressureSensor_Pump1 and Data.RelayActuator_Pump1 then
        local pressure = Data.PressureSensor_Pump1

        -- Automatic pump control based on pressure
        if pressure < 2.0 then
            Data.RelayActuator_Pump1 = true  -- Start pump
        elseif pressure > 8.0 then
            Data.RelayActuator_Pump1 = false -- Stop pump
        end

        -- High pressure alarm
        if pressure > PRESSURE_HIGH_LIMIT then
            if scan_count % 500 == 0 then
                plc_log(string.format("ALARM: Pump1 Pressure HIGH: %.2f bar", pressure))
                alarm_count = alarm_count + 1
            end
            -- Emergency: force pump OFF
            Data.RelayActuator_Pump1 = false
        end
    end

    -- ===================================================================
    -- Level Control Logic
    -- ===================================================================
    if Data.LevelSensor_Tank1 and Data.ValveActuator_Tank1 then
        local level = Data.LevelSensor_Tank1

        -- Level control: open valve when level is low
        if level < 30.0 then
            Data.ValveActuator_Tank1 = true  -- Open valve (fill)
        elseif level > 70.0 then
            Data.ValveActuator_Tank1 = false -- Close valve
        end

        -- Level alarms
        if level < 10.0 then
            if scan_count % 500 == 0 then
                plc_log(string.format("ALARM: Tank1 Level CRITICAL LOW: %.2f%%", level))
                alarm_count = alarm_count + 1
            end
        elseif level > 90.0 then
            if scan_count % 500 == 0 then
                plc_log(string.format("ALARM: Tank1 Level CRITICAL HIGH: %.2f%%", level))
                alarm_count = alarm_count + 1
            end
            -- Emergency: close valve
            Data.ValveActuator_Tank1 = false
        end
    end

    -- ===================================================================
    -- Vibration Monitoring
    -- ===================================================================
    if Data.VibrationSensor_Motor1 then
        local vib = Data.VibrationSensor_Motor1

        if vib > VIBRATION_ALARM_LEVEL then
            if scan_count % 200 == 0 then
                plc_log(string.format("ALARM: Motor1 Vibration HIGH: %.2f mm/s", vib))
                alarm_count = alarm_count + 1
            end
        end
    end

    if Data.VibrationSensor_Pump1 then
        local vib = Data.VibrationSensor_Pump1

        if vib > VIBRATION_ALARM_LEVEL then
            if scan_count % 200 == 0 then
                plc_log(string.format("ALARM: Pump1 Vibration HIGH: %.2f mm/s", vib))
                alarm_count = alarm_count + 1
            end
        end
    end

    -- ===================================================================
    -- Noise Level Monitoring
    -- ===================================================================
    if Data.NoiseSensor_FactoryFloor then
        local noise = Data.NoiseSensor_FactoryFloor

        if noise > 85.0 then
            if scan_count % 1000 == 0 then
                plc_log(string.format("WARNING: Factory noise level: %.1f dB", noise))
            end
        end
    end

    -- ===================================================================
    -- Motor Coordination Logic
    -- ===================================================================
    -- Conveyor and indexer coordination
    if Data.StepMotor_Conveyor and Data.StepMotor_Indexer then
        local conveyor_pos = Data.StepMotor_Conveyor
        local indexer_pos = Data.StepMotor_Indexer

        -- Example: When conveyor reaches certain position, trigger indexer
        -- (This is just demonstration - actual control would need more logic)
        if scan_count % 2000 == 0 then
            plc_log(string.format("Motor Status: Conveyor=%.1f steps, Indexer=%.1f steps",
                                  conveyor_pos, indexer_pos))
        end
    end

    -- ===================================================================
    -- Emergency Stop Logic
    -- ===================================================================
    if Data.EmergencyStop_Line1 then
        if Data.EmergencyStop_Line1 then
            -- Emergency stop pressed - halt all actuators
            if Data.RelayActuator_Pump1 then Data.RelayActuator_Pump1 = false end
            if Data.ValveActuator_Tank1 then Data.ValveActuator_Tank1 = false end
            if Data.RelayActuator_Heater then Data.RelayActuator_Heater = false end

            if scan_count % 500 == 0 then
                plc_log("EMERGENCY STOP ACTIVE - All actuators disabled")
            end
        end
    end

    -- ===================================================================
    -- Cooling Fan Control (based on temperature and spindle speed)
    -- ===================================================================
    if Data.FanSpeed_Cooling and Data.TemperatureSensor_Tank1 and Data.ServoMotor_Spindle then
        local temp = Data.TemperatureSensor_Tank1
        local spindle_speed = math.abs(Data.ServoMotor_Spindle or 0)

        -- Calculate required fan speed (0-100%)
        local temp_factor = map_value(temp, 20, 80, 0, 100)
        local spindle_factor = map_value(spindle_speed, 0, 3000, 0, 50)

        local required_fan_speed = clamp(temp_factor + spindle_factor, 0, 100)

        -- Smooth fan speed adjustment (ramp)
        local current_speed = Data.FanSpeed_Cooling
        local speed_diff = required_fan_speed - current_speed

        if math.abs(speed_diff) > 5 then
            if speed_diff > 0 then
                Data.FanSpeed_Cooling = math.min(current_speed + 2, required_fan_speed)
            else
                Data.FanSpeed_Cooling = math.max(current_speed - 2, required_fan_speed)
            end
        end
    end

    -- ===================================================================
    -- Heater Power Control (PID-like control)
    -- ===================================================================
    if Data.HeaterPower_Tank1 and Data.TemperatureSensor_Tank2 then
        local temp = Data.TemperatureSensor_Tank2
        local setpoint = 35.0
        local error = setpoint - temp

        -- Simple proportional control
        local kp = 5.0
        local power = 50 + (kp * error)  -- Base 50% + correction

        Data.HeaterPower_Tank1 = clamp(power, 0, 100)
    end
end

-- ===================================================================
-- Helper Functions
-- ===================================================================

-- Compare with tolerance
function fuzzy_equal(a, b, tolerance)
    tolerance = tolerance or 0.01
    return math.abs(a - b) < tolerance
end

-- Clamp value between min and max
function clamp(value, min_val, max_val)
    if value < min_val then return min_val end
    if value > max_val then return max_val end
    return value
end

-- Map value from one range to another
function map_value(value, in_min, in_max, out_min, out_max)
    return (value - in_min) * (out_max - out_min) / (in_max - in_min) + out_min
end

-- Rising edge detection
function rising_edge(current, previous)
    return current and not previous
end

-- Falling edge detection
function falling_edge(current, previous)
    return not current and previous
end

-- Hysteresis comparator
function hysteresis(value, threshold_high, threshold_low, previous_state)
    if value > threshold_high then
        return true
    elseif value < threshold_low then
        return false
    else
        return previous_state
    end
end
