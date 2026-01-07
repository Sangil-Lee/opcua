package plc

import (
	"fmt"
	"log"
	"os"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// LuaEngine manages Lua script execution for PLC logic
type LuaEngine struct {
	L           *lua.LState
	tagManager  *TagManager
	scriptPath  string
	scanTime    time.Duration
	running     bool
	stopChan    chan struct{}
	initialized bool
}

// NewLuaEngine creates a new Lua engine
func NewLuaEngine(tagManager *TagManager, scriptPath string, scanTimeMs int) *LuaEngine {
	return &LuaEngine{
		tagManager: tagManager,
		scriptPath: scriptPath,
		scanTime:   time.Duration(scanTimeMs) * time.Millisecond,
		stopChan:   make(chan struct{}),
	}
}

// Initialize initializes the Lua engine and loads the script
func (le *LuaEngine) Initialize() error {
	// Create new Lua state
	le.L = lua.NewState()

	// Check if script file exists
	if _, err := os.Stat(le.scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("Lua script file not found: %s", le.scriptPath)
	}

	// Register PLC functions
	le.registerPLCFunctions()

	// Create Data table with all tags
	le.createDataTable()

	// Load Lua script
	if err := le.L.DoFile(le.scriptPath); err != nil {
		return fmt.Errorf("failed to load Lua script: %w", err)
	}

	// Validate run_logic function exists
	runLogic := le.L.GetGlobal("run_logic")
	if runLogic.Type() != lua.LTFunction {
		return fmt.Errorf("Lua script must define 'run_logic()' function")
	}

	// Call init function if exists
	init := le.L.GetGlobal("init")
	if init.Type() == lua.LTFunction {
		if err := le.L.CallByParam(lua.P{
			Fn:      init,
			NRet:    0,
			Protect: true,
		}); err != nil {
			log.Printf("[LUA] Warning: init() function error: %v", err)
		} else {
			log.Println("[LUA] Initialization function executed successfully")
		}
	}

	le.initialized = true
	log.Printf("[LUA] Engine initialized successfully with script: %s", le.scriptPath)
	return nil
}

// registerPLCFunctions registers Go functions to Lua
func (le *LuaEngine) registerPLCFunctions() {
	// Get tag value
	le.L.SetGlobal("get_tag", le.L.NewFunction(func(L *lua.LState) int {
		tagName := L.CheckString(1)
		tag, err := le.tagManager.GetTag(tagName)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		value := tag.GetValue()
		switch v := value.(type) {
		case float64:
			L.Push(lua.LNumber(v))
		case int32:
			L.Push(lua.LNumber(v))
		case bool:
			L.Push(lua.LBool(v))
		case string:
			L.Push(lua.LString(v))
		default:
			L.Push(lua.LNil)
		}
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Set tag value
	le.L.SetGlobal("set_tag", le.L.NewFunction(func(L *lua.LState) int {
		tagName := L.CheckString(1)
		value := L.Get(2)

		tag, err := le.tagManager.GetTag(tagName)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		var goValue interface{}
		switch value.Type() {
		case lua.LTNumber:
			goValue = float64(value.(lua.LNumber))
		case lua.LTBool:
			goValue = bool(value.(lua.LBool))
		case lua.LTString:
			goValue = string(value.(lua.LString))
		default:
			L.Push(lua.LBool(false))
			L.Push(lua.LString("unsupported value type"))
			return 2
		}

		if err := tag.SetValue(goValue); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Log function
	le.L.SetGlobal("plc_log", le.L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		log.Printf("[PLC-LOGIC] %s", message)
		return 0
	}))

	// Get current time
	le.L.SetGlobal("get_time", le.L.NewFunction(func(L *lua.LState) int {
		now := time.Now()
		L.Push(lua.LNumber(now.Unix()))
		return 1
	}))

	// Sleep function (milliseconds)
	le.L.SetGlobal("sleep", le.L.NewFunction(func(L *lua.LState) int {
		ms := L.CheckInt(1)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return 0
	}))
}

// createDataTable creates the global Data table with all tags
func (le *LuaEngine) createDataTable() {
	dataTable := le.L.NewTable()

	tags := le.tagManager.GetAllTags()
	for _, tag := range tags {
		value := tag.GetValue()
		switch v := value.(type) {
		case float64:
			dataTable.RawSetString(tag.Name, lua.LNumber(v))
		case int32:
			dataTable.RawSetString(tag.Name, lua.LNumber(v))
		case bool:
			dataTable.RawSetString(tag.Name, lua.LBool(v))
		case string:
			dataTable.RawSetString(tag.Name, lua.LString(v))
		}
	}

	le.L.SetGlobal("Data", dataTable)
}

// UpdateDataTable updates the Data table with current tag values
func (le *LuaEngine) UpdateDataTable() {
	dataTable := le.L.GetGlobal("Data")
	if dataTable.Type() != lua.LTTable {
		log.Println("[LUA] Warning: Data table not found, recreating...")
		le.createDataTable()
		return
	}

	table := dataTable.(*lua.LTable)
	tags := le.tagManager.GetAllTags()

	for _, tag := range tags {
		value := tag.GetValue()
		switch v := value.(type) {
		case float64:
			table.RawSetString(tag.Name, lua.LNumber(v))
		case int32:
			table.RawSetString(tag.Name, lua.LNumber(v))
		case bool:
			table.RawSetString(tag.Name, lua.LBool(v))
		case string:
			table.RawSetString(tag.Name, lua.LString(v))
		}
	}
}

// SyncDataTableToTags syncs Data table values back to tag manager
func (le *LuaEngine) SyncDataTableToTags() error {
	dataTable := le.L.GetGlobal("Data")
	if dataTable.Type() != lua.LTTable {
		return fmt.Errorf("Data table not found")
	}

	table := dataTable.(*lua.LTable)
	tags := le.tagManager.GetAllTags()

	for _, tag := range tags {
		luaValue := table.RawGetString(tag.Name)

		var goValue interface{}
		switch luaValue.Type() {
		case lua.LTNumber:
			goValue = float64(luaValue.(lua.LNumber))
		case lua.LTBool:
			goValue = bool(luaValue.(lua.LBool))
		case lua.LTString:
			goValue = string(luaValue.(lua.LString))
		default:
			continue // Skip nil or unsupported types
		}

		if err := tag.SetValue(goValue); err != nil {
			log.Printf("[LUA] Warning: failed to sync tag %s: %v", tag.Name, err)
		}
	}

	return nil
}

// RunLogic executes the run_logic function once
func (le *LuaEngine) RunLogic() error {
	if !le.initialized {
		return fmt.Errorf("Lua engine not initialized")
	}

	// Update Data table with current tag values
	le.UpdateDataTable()

	// Get run_logic function
	runLogic := le.L.GetGlobal("run_logic")
	if runLogic.Type() != lua.LTFunction {
		return fmt.Errorf("run_logic function not found")
	}

	// Execute run_logic
	if err := le.L.CallByParam(lua.P{
		Fn:      runLogic,
		NRet:    0,
		Protect: true,
	}); err != nil {
		return fmt.Errorf("run_logic execution error: %w", err)
	}

	// Sync Data table back to tag manager
	if err := le.SyncDataTableToTags(); err != nil {
		return fmt.Errorf("failed to sync tags: %w", err)
	}

	return nil
}

// Start starts the PLC scan cycle
func (le *LuaEngine) Start() {
	if le.running {
		log.Println("[LUA] Engine already running")
		return
	}

	le.running = true
	log.Printf("[LUA] Starting PLC scan cycle (scan time: %v)", le.scanTime)

	go func() {
		ticker := time.NewTicker(le.scanTime)
		defer ticker.Stop()

		scanCount := 0
		for {
			select {
			case <-ticker.C:
				scanCount++
				if err := le.RunLogic(); err != nil {
					log.Printf("[LUA] Scan #%d error: %v", scanCount, err)
				}

				// Log every 100 scans
				if scanCount%100 == 0 {
					log.Printf("[LUA] Completed %d scan cycles", scanCount)
				}

			case <-le.stopChan:
				log.Println("[LUA] PLC scan cycle stopped")
				return
			}
		}
	}()
}

// Stop stops the PLC scan cycle
func (le *LuaEngine) Stop() {
	if !le.running {
		return
	}

	le.running = false
	close(le.stopChan)
	log.Println("[LUA] Stopping PLC scan cycle...")
}

// Close closes the Lua engine
func (le *LuaEngine) Close() {
	le.Stop()
	if le.L != nil {
		le.L.Close()
	}
	log.Println("[LUA] Engine closed")
}

// IsRunning returns whether the engine is running
func (le *LuaEngine) IsRunning() bool {
	return le.running
}

// GetScriptPath returns the script path
func (le *LuaEngine) GetScriptPath() string {
	return le.scriptPath
}
