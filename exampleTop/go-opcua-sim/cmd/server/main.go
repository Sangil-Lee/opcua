package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-opcua-sim/internal/config"
	"go-opcua-sim/internal/opcuaserver"
	"go-opcua-sim/internal/plc"
	"go-opcua-sim/internal/sim"
)

func main() {
	// Command line flags
	configFile := flag.String("config", "sensors.json", "Path to sensor configuration file")
	scriptFile := flag.String("script", "plc_logic.lua", "Path to PLC Lua script file")
	scanTimeMs := flag.Int("scantime", 100, "PLC scan time in milliseconds")
	enablePLC := flag.Bool("plc", true, "Enable PLC Lua logic execution")
	endpoint := flag.String("endpoint", "opc.tcp://0.0.0.0:4840", "OPC UA server endpoint")
	flag.Parse()

	fmt.Println("=== Go OPC UA PLC Simulation Server ===")

	// Load sensor configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load sensor config: %v", err)
	}
	fmt.Printf("[CONFIG] Loaded %d sensor definitions\n", len(cfg.Sensors))

	// Generate tags from sensor definitions
	tagManager, err := plc.GenerateTagsFromSensors(cfg.Sensors)
	if err != nil {
		log.Fatalf("Failed to generate tags: %v", err)
	}

	// Print tag summary
	plc.PrintTagSummary(tagManager)

	// Create sensor manager
	sensorManager, err := sim.NewSensorManager(tagManager, cfg)
	if err != nil {
		log.Fatalf("Failed to create sensor manager: %v", err)
	}

	// Start sensor simulation
	sensorManager.Start(100 * time.Millisecond)
	defer sensorManager.Stop()

	// Initialize PLC Lua Engine (if enabled)
	var luaEngine *plc.LuaEngine
	if *enablePLC {
		fmt.Println("\n=== Initializing PLC Lua Engine ===")

		// Create Lua engine
		luaEngine = plc.NewLuaEngine(tagManager, *scriptFile, *scanTimeMs)

		// Initialize Lua engine (loads script and calls init())
		if err := luaEngine.Initialize(); err != nil {
			log.Fatalf("Failed to initialize Lua engine: %v", err)
		}

		// Validate that run_logic function exists
		fmt.Println("[PLC] Validating Lua script...")
		if err := luaEngine.RunLogic(); err != nil {
			log.Fatalf("Failed to execute run_logic: %v\n"+
				"Make sure your Lua script defines a 'run_logic()' function", err)
		}
		fmt.Println("[PLC] Lua script validation successful")

		// Start PLC scan cycle
		luaEngine.Start()
		defer luaEngine.Close()

		fmt.Printf("[PLC] Lua engine started (scan time: %dms)\n", *scanTimeMs)
	}

	// Create and start OPC UA server
	opcuaServer := opcuaserver.NewOPCUAServer(*endpoint, tagManager)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := opcuaServer.Start(ctx); err != nil {
			log.Fatalf("OPC UA server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	fmt.Println("\nShutting down...")
	opcuaServer.Stop()
}
