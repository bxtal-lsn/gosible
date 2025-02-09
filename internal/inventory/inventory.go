package inventory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HostConfig stores per-host settings
type HostConfig struct {
	Host       string
	Group      string
	SSHUser    string
	SSHKeyFile string
	SSHPort    string
	Become     bool
}

// ✅ Function to create an inventory file with per-host settings
func CreateInventoryFile(directory string, hosts []HostConfig) string {
	// Ensure directory exists
	if err := os.MkdirAll(directory, 0o755); err != nil {
		fmt.Println("Error creating directory:", err)
		os.Exit(1)
	}

	// ✅ Generate unique inventory filename
	inventoryFile := getUniqueInventoryFilename(directory)

	// Write inventory content
	var inventoryContent strings.Builder
	inventoryContent.WriteString("---\nall:\n  hosts:\n")

	// ✅ Collect ungrouped hosts
	ungroupedHosts := []HostConfig{}
	groups := map[string][]HostConfig{}

	for _, host := range hosts {
		if host.Group == "" {
			ungroupedHosts = append(ungroupedHosts, host)
		} else {
			groups[host.Group] = append(groups[host.Group], host)
		}
	}

	// ✅ Write ungrouped hosts under `all: hosts`
	for _, host := range ungroupedHosts {
		inventoryContent.WriteString(fmt.Sprintf("    %s:\n", host.Host))
		inventoryContent.WriteString(fmt.Sprintf("      ansible_user: %s\n", host.SSHUser))
		inventoryContent.WriteString(fmt.Sprintf("      ansible_ssh_private_key_file: %s\n", host.SSHKeyFile))
		if host.SSHPort != "" {
			inventoryContent.WriteString(fmt.Sprintf("      ansible_port: %s\n", host.SSHPort))
		}
		if host.Become {
			inventoryContent.WriteString("      ansible_become: true\n")
		}
	}

	// ✅ Write grouped hosts under a **single `children:` section**
	if len(groups) > 0 {
		inventoryContent.WriteString("\n  children:\n")
		for groupName, groupHosts := range groups {
			inventoryContent.WriteString(fmt.Sprintf("    %s:\n      hosts:\n", groupName))
			for _, host := range groupHosts {
				inventoryContent.WriteString(fmt.Sprintf("        %s:\n", host.Host))
				inventoryContent.WriteString(fmt.Sprintf("          ansible_user: %s\n", host.SSHUser))
				inventoryContent.WriteString(fmt.Sprintf("          ansible_ssh_private_key_file: %s\n", host.SSHKeyFile))
				if host.SSHPort != "" {
					inventoryContent.WriteString(fmt.Sprintf("          ansible_port: %s\n", host.SSHPort))
				}
				if host.Become {
					inventoryContent.WriteString("          ansible_become: true\n")
				}
			}
		}
	}

	// Save inventory file
	err := os.WriteFile(inventoryFile, []byte(inventoryContent.String()), 0o644)
	if err != nil {
		fmt.Println("Error writing inventory file:", err)
		os.Exit(1)
	}

	return inventoryFile
}

// ✅ Function to generate a unique filename if `inventory.yml` exists
func getUniqueInventoryFilename(directory string) string {
	baseName := "inv"
	ext := ".yml"
	filename := filepath.Join(directory, baseName+ext)

	// ✅ Check if file exists and increment name if necessary
	counter := 1
	for fileExists(filename) {
		filename = filepath.Join(directory, fmt.Sprintf("%s%d%s", baseName, counter, ext))
		counter++
	}

	return filename
}

// ✅ Function to check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
