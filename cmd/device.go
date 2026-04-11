package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/api"
	"github.com/tcl-eaglelab/tcl-smart-home-cli/internal/output"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device query and management",
}

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonFlag, _ := cmd.Flags().GetBool("json")

		client, err := api.NewClient()
		if err != nil {
			return err
		}

		devices, err := client.GetAllDevices()
		if err != nil {
			return err
		}

		if jsonFlag {
			data, err := json.MarshalIndent(devices, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		if len(devices) == 0 {
			output.Gray("No devices found.")
			return nil
		}

		output.Gray(fmt.Sprintf("Total: %d device(s)", len(devices)))
		fmt.Println()

		rows := make([][]string, len(devices))
		for i, d := range devices {
			familyName := "-"
			if d.Family != nil && d.Family.FamilyName != "" {
				familyName = d.Family.FamilyName
			}
			rows[i] = []string{
				d.DeviceID,
				nonEmpty(d.NickName),
				nonEmpty(d.Category),
				familyName,
				nonEmpty(d.LocationName),
				output.StatusBadge(d.IsDeviceOnline()),
			}
		}
		output.Table([]string{"Device ID", "Name", "Category", "Home", "Location", "Status"}, rows)
		return nil
	},
}

var deviceDetailCmd = &cobra.Command{
	Use:   "detail <device_id>",
	Short: "Get device detail and current status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID := args[0]
		jsonFlag, _ := cmd.Flags().GetBool("json")

		client, err := api.NewClient()
		if err != nil {
			return err
		}

		device, err := findDevice(client, deviceID)
		if err != nil {
			return err
		}

		status, err := client.GetDeviceStatus(deviceID)
		if err != nil {
			return err
		}

		if jsonFlag {
			result := map[string]interface{}{"device": device, "status": status}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		familyName := "-"
		if device.Family != nil && device.Family.FamilyName != "" {
			familyName = device.Family.FamilyName
		}

		fmt.Println()
		output.Bold("  " + nonEmpty(device.NickName))
		output.Gray("  ─────────────────────────────")
		fmt.Printf("  ID:        %s\n", device.DeviceID)
		fmt.Printf("  Category:  %s\n", nonEmpty(device.Category))
		fmt.Printf("  Home:      %s\n", familyName)
		fmt.Printf("  Location:  %s\n", nonEmpty(device.LocationName))
		fmt.Printf("  Status:    %s\n", output.StatusBadge(device.IsDeviceOnline()))

		if len(status) > 0 {
			fmt.Println()
			output.Bold("  Properties:")
			keys := make([]string, 0, len(status))
			for k := range status {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			rows := make([][]string, 0, len(status))
			for _, k := range keys {
				rows = append(rows, []string{k, fmt.Sprintf("%v", status[k])})
			}
			output.Table([]string{"Code", "Value"}, rows)
		}

		return nil
	},
}

var deviceModelCmd = &cobra.Command{
	Use:   "model <device_id>",
	Short: "Get device Thing Model (supported capabilities)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID := args[0]
		jsonFlag, _ := cmd.Flags().GetBool("json")

		client, err := api.NewClient()
		if err != nil {
			return err
		}

		device, err := findDevice(client, deviceID)
		if err != nil {
			return err
		}
		if device.ProductKey == "" {
			return fmt.Errorf("device has no productKey, cannot fetch Thing Model")
		}

		model, err := client.GetDeviceModel(device.ProductKey)
		if err != nil {
			return err
		}

		if jsonFlag {
			data, err := json.MarshalIndent(model, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		props := model.Properties
		if len(props) == 0 {
			output.Gray("No model data found.")
			return nil
		}

		fmt.Println()
		output.Bold(fmt.Sprintf("  Properties (%d):", len(props)))
		rows := make([][]string, len(props))
		for i, p := range props {
			rows[i] = []string{
				p.Identifier,
				nonEmpty(p.Name),
				nonEmpty(p.AccessMode),
				p.DataType.Type,
				formatSpecs(p.DataType),
			}
		}
		output.Table([]string{"Code", "Name", "Access", "Type", "Spec"}, rows)
		return nil
	},
}

var deviceControlCmd = &cobra.Command{
	Use:   "control <device_id> <properties>",
	Short: "Control a device by issuing properties",
	Long:  `Control a device by issuing properties. Example: tcl device control 12345 '{"powerSwitch":1}'`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceID := args[0]
		propertiesStr := args[1]
		jsonFlag, _ := cmd.Flags().GetBool("json")

		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(propertiesStr), &properties); err != nil {
			output.Error(fmt.Sprintf("Invalid JSON: %v\nExample: '{\"powerSwitch\":1}'", err))
			return errSilent
		}
		if len(properties) == 0 {
			output.Error("No properties specified. Example: '{\"powerSwitch\":1}'")
			return errSilent
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}

		device, err := findDevice(client, deviceID)
		if err != nil {
			return err
		}

		if !device.IsDeviceOnline() {
			output.Error("Device is offline, cannot send control command. Check device WiFi connection.")
			return errSilent
		}

		var readonlyKeys []string
		if device.ProductKey != "" {
			model, err := client.GetDeviceModel(device.ProductKey)
			if err != nil {
				return fmt.Errorf("failed to fetch Thing Model for safety check: %w", err)
			}
			if len(model.Properties) > 0 {
				writable := make(map[string]interface{})

				for key, val := range properties {
					isReadonly := false
					for _, p := range model.Properties {
						if p.Identifier == key && p.AccessMode == "r" {
							isReadonly = true
							break
						}
					}
					if isReadonly {
						readonlyKeys = append(readonlyKeys, key)
					} else {
						writable[key] = val
					}
				}

				if len(readonlyKeys) > 0 {
					sort.Strings(readonlyKeys)
					if !jsonFlag {
						output.Warn("Read-only properties filtered out: " + strings.Join(readonlyKeys, ", "))
					}
					if len(writable) == 0 {
						output.Error("No writable properties to send.")
						return errSilent
					}
					properties = writable
				}
			}
		}

		_, err = client.ControlDevice(deviceID, properties)
		if err != nil {
			return err
		}

		if jsonFlag {
			result := map[string]interface{}{"success": true, "deviceId": deviceID, "properties": properties}
			if len(readonlyKeys) > 0 {
				result["filtered"] = readonlyKeys
			}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		data, err := json.Marshal(properties)
		if err != nil {
			return fmt.Errorf("failed to marshal properties: %w", err)
		}
		output.Success("Command sent to device " + deviceID)
		output.Gray("  Properties: " + string(data))
		output.Gray("  Run 'tcl device detail " + deviceID + "' to verify the current state")
		return nil
	},
}

// findDevice fetches all devices and matches by ID.
// TCL IoT API does not provide a single-device query endpoint, so a full list query is required.
func findDevice(client *api.Client, deviceID string) (*api.Device, error) {
	devices, err := client.GetAllDevices()
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if devices[i].DeviceID == deviceID {
			return &devices[i], nil
		}
	}
	return nil, fmt.Errorf("device not found: %s (use \"tcl device list\" to see available devices)", deviceID)
}

// formatSpecs formats the specs field for human-readable table display.
// Tries to extract min/max/step/unit for numeric types; falls back to type name.
func formatSpecs(dt api.DataType) string {
	if len(dt.Specs) == 0 {
		return nonEmpty(dt.Type)
	}
	var m map[string]string
	if err := json.Unmarshal(dt.Specs, &m); err == nil {
		if min, ok := m["min"]; ok {
			if max, ok2 := m["max"]; ok2 {
				s := min + "~" + max
				if step, ok3 := m["step"]; ok3 {
					s += " step:" + step
				}
				if unit, ok4 := m["unit"]; ok4 {
					s += " " + unit
				}
				return s
			}
		}
	}
	return nonEmpty(dt.Type)
}

func nonEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func init() {
	rootCmd.AddCommand(deviceCmd)

	deviceCmd.PersistentFlags().Bool("json", false, "Output as JSON")

	deviceCmd.AddCommand(deviceListCmd)
	deviceCmd.AddCommand(deviceDetailCmd)
	deviceCmd.AddCommand(deviceModelCmd)
	deviceCmd.AddCommand(deviceControlCmd)
}
