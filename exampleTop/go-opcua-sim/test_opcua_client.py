#!/usr/bin/env python3
"""
OPC UA Client Test Script
Tests connection to go-opcua-sim server
"""

import sys
import time

try:
    from opcua import Client, ua
except ImportError:
    print("Error: python-opcua library not installed")
    print("Install with: pip install opcua")
    sys.exit(1)


def test_connection(endpoint="opc.tcp://127.0.0.1:4840"):
    """Test basic connection to OPC UA server"""
    print("=" * 60)
    print("OPC UA Client Test")
    print("=" * 60)
    print(f"\nConnecting to: {endpoint}")

    client = Client(endpoint)
    client.set_security_string("None")

    try:
        # Try to connect (with timeout)
        client.connect()
        print("✓ Connected successfully!")

        # Get server info
        print("\n--- Server Information ---")
        print(f"Server URI: {client.get_server_node()}")

        # Browse root
        print("\n--- Browsing Root Node ---")
        root = client.get_root_node()
        print(f"Root: {root}")

        # Try to read a specific node (TemperatureSensor_Tank1)
        print("\n--- Reading Temperature Sensor ---")
        node_id = "ns=2;i=1000"  # TemperatureSensor_Tank1
        print(f"Node ID: {node_id}")

        temp_node = client.get_node(node_id)
        value = temp_node.get_value()
        print(f"✓ Temperature: {value}°C")

        # Read multiple values
        print("\n--- Reading Multiple Sensors ---")
        sensors = [
            ("ns=2;i=1000", "TemperatureSensor_Tank1"),
            ("ns=2;i=1001", "LevelSensor_Tank1"),
            ("ns=2;i=1002", "RelayActuator_Pump1"),
            ("ns=2;i=1003", "MotorSpeed_Conveyor"),
        ]

        for node_id, name in sensors:
            try:
                node = client.get_node(node_id)
                value = node.get_value()
                print(f"  {name}: {value}")
            except Exception as e:
                print(f"  {name}: Error - {e}")

        # Continuous monitoring
        print("\n--- Continuous Monitoring (5 seconds) ---")
        print("Press Ctrl+C to stop")

        for i in range(5):
            temp_node = client.get_node("ns=2;i=1000")
            value = temp_node.get_value()
            timestamp = time.strftime("%H:%M:%S")
            print(f"[{timestamp}] Temperature: {value:.2f}°C")
            time.sleep(1)

        print("\n✓ Test completed successfully!")

    except Exception as e:
        print(f"\n✗ Connection failed: {e}")
        print("\nPossible reasons:")
        print("  1. Server is not running")
        print("  2. Server network listener not implemented")
        print("  3. Firewall blocking connection")
        print("  4. Wrong endpoint address")

    finally:
        client.disconnect()
        print("\nConnection closed.")


def test_with_timeout(endpoint="opc.tcp://localhost:4840", timeout=5):
    """Test connection with timeout"""
    print("=" * 60)
    print("OPC UA Client Test (with timeout)")
    print("=" * 60)
    print(f"\nConnecting to: {endpoint}")
    print(f"Timeout: {timeout} seconds")

    client = Client(endpoint)
    client.set_security_string("None")
    client.session_timeout = timeout * 1000  # milliseconds

    try:
        client.connect()
        print("✓ Connected!")

        # Quick test
        node = client.get_node("ns=2;i=1000")
        value = node.get_value()
        print(f"✓ Read value: {value}")

        client.disconnect()
        return True

    except Exception as e:
        print(f"✗ Failed: {e}")
        return False


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="OPC UA Client Test")
    parser.add_argument(
        "-e",
        "--endpoint",
        default="opc.tcp://localhost:4840",
        help="OPC UA server endpoint",
    )
    parser.add_argument(
        "-t", "--timeout", type=int, default=5, help="Connection timeout (seconds)"
    )
    parser.add_argument("--quick", action="store_true", help="Quick test with timeout")

    args = parser.parse_args()

    if args.quick:
        success = test_with_timeout(args.endpoint, args.timeout)
        sys.exit(0 if success else 1)
    else:
        test_connection(args.endpoint)
