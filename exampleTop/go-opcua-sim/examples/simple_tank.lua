-- Simple Tank Temperature Control Example
-- This script demonstrates basic PLC logic for temperature control

-- ===================================================================
-- Global Variables (persist across scans)
-- ===================================================================
scan_count = scan_count or 0

-- ===================================================================
-- Initialization Function (called once at startup)
-- ===================================================================
function init()
    plc_log("========================================")
    plc_log("Simple Tank Control System Initialized")
    plc_log("========================================")
    plc_log("Target Temperature: 25°C")
    plc_log("Hysteresis: ±2°C")
    plc_log("----------------------------------------")
end

-- ===================================================================
-- Main Logic Function (called every scan cycle)
-- ===================================================================
function run_logic()
    scan_count = scan_count + 1

    -- Read sensors
    local temp = Data.TankTemperature
    local level = Data.TankLevel

    -- Temperature control with hysteresis
    -- Turn heater ON if temp < 23°C
    -- Turn heater OFF if temp > 27°C
    if temp < 23 then
        Data.Heater = true
    elseif temp > 27 then
        Data.Heater = false
    end
    -- Between 23-27°C: maintain current state

    -- Level control
    -- Turn pump ON if level < 40%
    -- Turn pump OFF if level > 60%
    if level < 40 then
        Data.Pump = true
    elseif level > 60 then
        Data.Pump = false
    end

    -- Log status every 100 scans (every 10 seconds at 100ms scan time)
    if scan_count % 100 == 0 then
        plc_log(string.format("Status Report #%d:", scan_count / 100))
        plc_log(string.format("  Temperature: %.2f°C", temp))
        plc_log(string.format("  Level: %.1f%%", level))
        plc_log(string.format("  Heater: %s", Data.Heater and "ON" or "OFF"))
        plc_log(string.format("  Pump: %s", Data.Pump and "ON" or "OFF"))
        plc_log("----------------------------------------")
    end
end
