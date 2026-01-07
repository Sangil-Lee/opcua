package opcuaserver

import (
	"context"
	"fmt"
	"go-opcua-sim/internal/plc"
	"log"
	"sync"
	"time"

	"github.com/awcullen/opcua/server"
	"github.com/awcullen/opcua/ua"
)

// OPCUAServer wraps the awcullen OPC UA server
type OPCUAServer struct {
	endpoint    string
	tagManager  *plc.TagManager
	ctx         context.Context
	cancel      context.CancelFunc
	nodeMapping map[string]string // tag name -> node ID string
	server      *server.Server
	mu          sync.RWMutex
	running     bool
}

// NewOPCUAServer creates a new OPC UA server
func NewOPCUAServer(endpoint string, tagManager *plc.TagManager) *OPCUAServer {
	return &OPCUAServer{
		endpoint:    endpoint,
		tagManager:  tagManager,
		nodeMapping: make(map[string]string),
	}
}

// Start starts the OPC UA server
func (s *OPCUAServer) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	log.Printf("[OPCUA] Starting OPC UA server at %s", s.endpoint)
	log.Printf("[OPCUA] Available tags: %d", s.tagManager.GetTagCount())

	// Create server instance
	srv, err := server.New(
		ua.ApplicationDescription{
			ApplicationURI: "urn:go-opcua-sim",
			ProductURI:     "urn:go-opcua-sim",
			ApplicationName: ua.LocalizedText{
				Text:   "Go OPC UA Simulator",
				Locale: "en",
			},
			ApplicationType: ua.ApplicationTypeServer,
		},
		"./pki/server.crt",
		"./pki/server.key",
		s.endpoint,
		server.WithBuildInfo(ua.BuildInfo{
			ProductName:      "Go OPC UA Simulator",
			SoftwareVersion:  "1.0.0",
			ManufacturerName: "go-opcua-sim",
		}),
		server.WithAnonymousIdentity(true),
		server.WithSecurityPolicyNone(true),
		server.WithInsecureSkipVerify(),
	)
	if err != nil {
		return fmt.Errorf("failed to create OPC UA server: %v", err)
	}
	s.server = srv

	// Register all tag nodes
	if err := s.registerNodes(); err != nil {
		return fmt.Errorf("failed to register nodes: %v", err)
	}

	// Start update goroutine to sync tag values to OPC UA nodes
	go s.updateNodeValues()

	// Start the server
	log.Printf("[OPCUA] Server listening on %s", s.endpoint)
	log.Printf("[OPCUA] Server is ready to accept connections")

	s.running = true

	// Run server in a goroutine (non-blocking)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != ua.BadServerHalted {
			log.Printf("[OPCUA] Server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-s.ctx.Done()
	return nil
}

// registerNodes registers all tag nodes in the OPC UA server
func (s *OPCUAServer) registerNodes() error {
	tags := s.tagManager.GetAllTags()
	nm := s.server.NamespaceManager()

	fmt.Println("\n=== OPC UA Server Tag Mapping ===")
	fmt.Printf("%-40s %-50s %s\n", "Tag Name", "NodeID", "Data Type")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	// Find or get the Objects folder (standard namespace 0, id 85)
	objectsFolderNodeID := ua.ParseNodeID("i=85")

	var nodesToAdd []server.Node

	for _, tag := range tags {
		// Use tag name as string identifier
		nodeIDString := tag.Name
		s.nodeMapping[tag.Name] = nodeIDString

		// Determine OPC UA data type and initial value
		var dataType ua.NodeID
		var initialValue ua.DataValue

		switch tag.Type {
		case plc.TagTypeBool:
			dataType = ua.DataTypeIDBoolean
			v, _ := s.tagManager.GetTagValue(tag.Name)
			if boolVal, ok := v.(bool); ok {
				initialValue = ua.NewDataValue(boolVal, 0, time.Now(), 0, time.Now(), 0)
			} else {
				initialValue = ua.NewDataValue(false, 0, time.Now(), 0, time.Now(), 0)
			}

		case plc.TagTypeInt32:
			dataType = ua.DataTypeIDInt32
			v, _ := s.tagManager.GetTagValue(tag.Name)
			if intVal, ok := v.(int32); ok {
				initialValue = ua.NewDataValue(intVal, 0, time.Now(), 0, time.Now(), 0)
			} else {
				initialValue = ua.NewDataValue(int32(0), 0, time.Now(), 0, time.Now(), 0)
			}

		case plc.TagTypeFloat64:
			dataType = ua.DataTypeIDDouble
			v, _ := s.tagManager.GetTagValue(tag.Name)
			if floatVal, ok := v.(float64); ok {
				initialValue = ua.NewDataValue(floatVal, 0, time.Now(), 0, time.Now(), 0)
			} else {
				initialValue = ua.NewDataValue(float64(0.0), 0, time.Now(), 0, time.Now(), 0)
			}

		default:
			dataType = ua.DataTypeIDDouble
			initialValue = ua.NewDataValue(float64(0.0), 0, time.Now(), 0, time.Now(), 0)
		}

		// Create variable node with string identifier
		varNode := server.NewVariableNode(
			s.server,
			ua.NodeIDString{NamespaceIndex: 2, ID: nodeIDString},
			ua.QualifiedName{
				NamespaceIndex: 2,
				Name:           tag.Name,
			},
			ua.LocalizedText{
				Text: tag.Name,
			},
			ua.LocalizedText{
				Text: tag.Description,
			},
			nil,
			[]ua.Reference{
				{
					ReferenceTypeID: ua.ReferenceTypeIDOrganizes,
					IsInverse:       true,
					TargetID:        ua.ExpandedNodeID{NodeID: objectsFolderNodeID},
				},
			},
			initialValue,
			dataType,
			ua.ValueRankScalar,
			[]uint32{},
			ua.AccessLevelsCurrentRead|ua.AccessLevelsCurrentWrite,
			250.0,
			false,
			nil,
		)

		nodesToAdd = append(nodesToAdd, varNode)

		dataTypeStr := "Double"
		if tag.Type == plc.TagTypeBool {
			dataTypeStr = "Boolean"
		} else if tag.Type == plc.TagTypeInt32 {
			dataTypeStr = "Int32"
		}

		fmt.Printf("%-40s %-50s %s\n", tag.Name, fmt.Sprintf("ns=2;s=%s", nodeIDString), dataTypeStr)
	}

	// Add all nodes at once
	nm.AddNodes(nodesToAdd...)

	fmt.Println()
	return nil
}

// updateNodeValues continuously updates OPC UA node values from tag manager
func (s *OPCUAServer) updateNodeValues() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	nm := s.server.NamespaceManager()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if !s.running {
				return
			}

			s.mu.RLock()
			for tagName, nodeIDStr := range s.nodeMapping {
				// Get current value from tag manager
				value, err := s.tagManager.GetTagValue(tagName)
				if err != nil {
					continue
				}

				// Find the node using string identifier
				nodeIDObj := ua.ParseNodeID(fmt.Sprintf("ns=2;s=%s", nodeIDStr))
				if varNode, ok := nm.FindVariable(nodeIDObj); ok {
					// Create new data value
					var dataValue ua.DataValue

					switch v := value.(type) {
					case float64:
						dataValue = ua.NewDataValue(v, 0, time.Now(), 0, time.Now(), 0)
					case int32:
						dataValue = ua.NewDataValue(v, 0, time.Now(), 0, time.Now(), 0)
					case bool:
						dataValue = ua.NewDataValue(v, 0, time.Now(), 0, time.Now(), 0)
					case string:
						dataValue = ua.NewDataValue(v, 0, time.Now(), 0, time.Now(), 0)
					default:
						continue
					}

					varNode.SetValue(dataValue)
				}
			}
			s.mu.RUnlock()
		}
	}
}

// Stop stops the OPC UA server
func (s *OPCUAServer) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	if s.server != nil {
		s.server.Close()
	}

	if s.cancel != nil {
		s.cancel()
	}

	log.Println("[OPCUA] Server stopped")
}

// GetNodeID returns the node ID string for a tag name
func (s *OPCUAServer) GetNodeID(tagName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodeID, ok := s.nodeMapping[tagName]
	if !ok {
		return "", fmt.Errorf("tag '%s' not found in node mapping", tagName)
	}
	return nodeID, nil
}

// GetNodeIDString returns the full node ID string for a tag name (ns=2;s=...)
func (s *OPCUAServer) GetNodeIDString(tagName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodeID, ok := s.nodeMapping[tagName]
	if !ok {
		return "", fmt.Errorf("tag '%s' not found in node mapping", tagName)
	}
	return fmt.Sprintf("ns=2;s=%s", nodeID), nil
}

// ReadTagValue reads a tag value by node ID string
func (s *OPCUAServer) ReadTagValue(nodeIDStr string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find tag by node ID
	for tagName, nid := range s.nodeMapping {
		if nid == nodeIDStr {
			return s.tagManager.GetTagValue(tagName)
		}
	}
	return nil, fmt.Errorf("node ID 'ns=2;s=%s' not found", nodeIDStr)
}

// WriteTagValue writes a value to a tag by node ID string
func (s *OPCUAServer) WriteTagValue(nodeIDStr string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find tag by node ID
	for tagName, nid := range s.nodeMapping {
		if nid == nodeIDStr {
			return s.tagManager.SetTagValue(tagName, value)
		}
	}
	return fmt.Errorf("node ID 'ns=2;s=%s' not found", nodeIDStr)
}

// GetAllNodeValues returns all node values
func (s *OPCUAServer) GetAllNodeValues() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for tagName := range s.nodeMapping {
		if value, err := s.tagManager.GetTagValue(tagName); err == nil {
			result[tagName] = value
		}
	}
	return result
}
