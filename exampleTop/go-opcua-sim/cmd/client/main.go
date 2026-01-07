package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func main() {
	endpoint := flag.String("endpoint", "opc.tcp://localhost:4840", "OPC UA server endpoint")
	nodeID := flag.String("node", "ns=2;i=1000", "Node ID to read (e.g., ns=2;i=1000)")
	continuous := flag.Bool("continuous", false, "Continuously read values")
	interval := flag.Int("interval", 1000, "Read interval in milliseconds (for continuous mode)")
	flag.Parse()

	fmt.Println("=== OPC UA Client Test ===")
	fmt.Printf("Connecting to: %s\n", *endpoint)

	ctx := context.Background()

	// Create OPC UA client
	client, err := opcua.NewClient(*endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Connect to server
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close(ctx)

	fmt.Println("Connected successfully!")

	// Parse node ID
	nid, err := ua.ParseNodeID(*nodeID)
	if err != nil {
		log.Fatalf("Failed to parse node ID: %v", err)
	}

	if *continuous {
		// Continuous reading mode
		fmt.Printf("Reading node %s every %d ms (Press Ctrl+C to stop)\n", *nodeID, *interval)
		ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			readAndPrintValue(ctx, client, nid)
		}
	} else {
		// Single read mode
		fmt.Printf("Reading node: %s\n", *nodeID)
		readAndPrintValue(ctx, client, nid)
	}
}

func readAndPrintValue(ctx context.Context, client *opcua.Client, nodeID *ua.NodeID) {
	// Create read request
	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			{NodeID: nodeID},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	// Read value
	resp, err := client.Read(ctx, req)
	if err != nil {
		log.Printf("Read failed: %v", err)
		return
	}

	// Check results
	if len(resp.Results) == 0 {
		log.Println("No results returned")
		return
	}

	result := resp.Results[0]
	if result.Status != ua.StatusOK {
		log.Printf("Read error: %v", result.Status)
		return
	}

	// Print value
	timestamp := result.ServerTimestamp.Format("2006-01-02 15:04:05.000")
	fmt.Printf("[%s] Value: %v (Type: %T)\n", timestamp, result.Value.Value(), result.Value.Value())
}
