package plc

import (
	"fmt"
	"go-opcua-sim/internal/config"
	"log"
	"strings"
)

// GenerateTagsFromSensors creates tags from sensor definitions
func GenerateTagsFromSensors(sensorDefs []config.SensorDefinition) (*TagManager, error) {
	tagManager := NewTagManager()

	for _, sensor := range sensorDefs {
		tagType := determineTagType(sensor.Address, sensor.Type)

		tag := NewTag(
			sensor.Name,
			sensor.Address,
			sensor.Description,
			tagType,
		)

		if err := tagManager.AddTag(tag); err != nil {
			return nil, fmt.Errorf("failed to add tag '%s': %w", sensor.Name, err)
		}
	}

	log.Printf("[TAG] Generated %d tags from sensor definitions", tagManager.GetTagCount())
	return tagManager, nil
}

// determineTagType determines the tag type based on address and sensor type
func determineTagType(address, sensorType string) TagType {
	// Check address prefix
	if strings.HasPrefix(address, "%DF") {
		// Double Float - all analog sensors and motors
		return TagTypeFloat64
	} else if strings.HasPrefix(address, "%MW") {
		// Memory Word (bit) - digital sensors and relays
		return TagTypeBool
	} else if strings.HasPrefix(address, "%DW") {
		// Data Word - integer actuators
		return TagTypeInt32
	}

	// Fallback: determine by sensor type
	switch sensorType {
	case "temperature", "pressure", "sine", "random", "vibration", "noise":
		return TagTypeFloat64
	case "stepmotor", "servomotor":
		return TagTypeFloat64
	case "digital", "relay":
		return TagTypeBool
	case "integer":
		return TagTypeInt32
	default:
		return TagTypeFloat64 // Default to float64
	}
}

// PrintTagSummary prints a summary of all tags
func PrintTagSummary(tagManager *TagManager) {
	tags := tagManager.GetAllTags()

	fmt.Println("\n=== PLC Tag Summary ===")
	fmt.Printf("Total Tags: %d\n\n", len(tags))

	typeCount := make(map[TagType]int)
	for _, tag := range tags {
		typeCount[tag.Type]++
	}

	fmt.Println("By Type:")
	if count, ok := typeCount[TagTypeFloat64]; ok {
		fmt.Printf("  Float64: %d tags\n", count)
	}
	if count, ok := typeCount[TagTypeInt32]; ok {
		fmt.Printf("  Int32:   %d tags\n", count)
	}
	if count, ok := typeCount[TagTypeBool]; ok {
		fmt.Printf("  Bool:    %d tags\n", count)
	}
	if count, ok := typeCount[TagTypeString]; ok {
		fmt.Printf("  String:  %d tags\n", count)
	}

	fmt.Println("\nTag List:")
	fmt.Printf("%-40s %-12s %-10s %s\n", "Name", "Address", "Type", "Description")
	fmt.Println(strings.Repeat("-", 100))

	for _, tag := range tags {
		typeStr := getTagTypeString(tag.Type)
		desc := tag.Description
		if len(desc) > 35 {
			desc = desc[:32] + "..."
		}
		fmt.Printf("%-40s %-12s %-10s %s\n", tag.Name, tag.Address, typeStr, desc)
	}
	fmt.Println()
}

// getTagTypeString returns string representation of TagType
func getTagTypeString(tagType TagType) string {
	switch tagType {
	case TagTypeFloat64:
		return "Float64"
	case TagTypeInt32:
		return "Int32"
	case TagTypeBool:
		return "Bool"
	case TagTypeString:
		return "String"
	default:
		return "Unknown"
	}
}

// GenerateLuaTemplate generates a template Lua script based on tags
func GenerateLuaTemplate(tagManager *TagManager) string {
	tags := tagManager.GetAllTags()

	template := `-- PLC Logic Script (Auto-generated template)
-- This script is executed periodically by the PLC engine

-- ===================================================================
-- Available Tags (from sensors.json)
-- ===================================================================
--[[
`

	for _, tag := range tags {
		typeStr := getTagTypeString(tag.Type)
		template += fmt.Sprintf("  Data.%-35s -- %s (%s, %s)\n",
			tag.Name,
			tag.Description,
			tag.Address,
			typeStr)
	}

	template += `]]

-- ===================================================================
-- Global Variables (persist across scans)
-- ===================================================================
scan_count = scan_count or 0
last_temp = last_temp or 0

-- ===================================================================
-- Initialization Function (called once at startup)
-- ===================================================================
function init()
    plc_log("PLC logic initialized")
    plc_log("Total tags: " .. #Data)
end

-- ===================================================================
-- Main Logic Function (called every scan cycle)
-- ===================================================================
function run_logic()
    -- Increment scan counter
    scan_count = scan_count + 1

    -- Example: Log every 100 scans
    if scan_count % 100 == 0 then
        plc_log(string.format("Scan #%d", scan_count))
    end

    -- Example: Read sensor values
    -- local temp = Data.TemperatureSensor_Tank1
    -- if temp then
    --     plc_log(string.format("Tank1 Temperature: %.2f", temp))
    -- end

    -- Example: Simple logic
    -- if Data.TemperatureSensor_Tank1 > 50 then
    --     Data.RelayActuator_Heater = false  -- Turn off heater
    -- else
    --     Data.RelayActuator_Heater = true   -- Turn on heater
    -- end

    -- Add your custom logic here...
end

-- ===================================================================
-- Helper Functions
-- ===================================================================

-- Compare with tolerance
function fuzzy_equal(a, b, tolerance)
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
`

	return template
}
