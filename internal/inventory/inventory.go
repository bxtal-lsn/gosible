package inventory

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ✅ HostConfig stores per-host settings
type HostConfig struct {
	Host       string
	Group      string
	SSHUser    string
	SSHKeyFile string
	SSHPort    string
	Become     bool
}

// ✅ Define an overridable `execCommand` function for testing
var execCommand = exec.Command

// ✅ Function to create an inventory file with per-host settings
func CreateInventoryFile(directory string, hosts []HostConfig) (string, error) {
	// Ensure directory exists
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
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

	// ✅ Write grouped hosts under `children:` (fixed recursive children issue)
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

	// ✅ Save inventory file
	err := os.WriteFile(inventoryFile, []byte(inventoryContent.String()), 0o644)
	if err != nil {
		return "", fmt.Errorf("error writing inventory file: %w", err)
	}

	return inventoryFile, nil
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

// ✅ Auto-discover Multipass/Docker instances (NOW MOCKABLE!)
// ✅ Discover running Multipass/Docker instances (WITH TESTABLE INPUT)
func DiscoverInstances(reader *bufio.Reader) []string {
	var instances []string

	// ✅ Detect Multipass instances
	fmt.Println("\n🔍 Checking for running Multipass instances...")
	out, err := execCommand("multipass", "list", "--format", "csv").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines[1:] { // Skip header row
			fields := strings.Split(line, ",")
			if len(fields) > 2 && strings.TrimSpace(fields[1]) == "Running" {
				instances = append(instances, strings.TrimSpace(fields[2])) // Extract IP
			}
		}
	}

	// ✅ Detect Docker containers
	fmt.Println("\n🐳 Checking for running Docker containers...")
	out, err = execCommand("docker", "ps", "--format", "{{.Names}}").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if len(line) > 0 {
				instances = append(instances, strings.TrimSpace(line)) // Use container name
			}
		}
	}

	// ✅ Return all instances if none are found
	if len(instances) == 0 {
		fmt.Println("⚠️ No running Multipass or Docker instances found.")
		return []string{}
	}

	// ✅ Prompt user for instance selection (allow input override in tests)
	fmt.Println("\n🔍 Found the following instances:")
	for i, instance := range instances {
		fmt.Printf("[%d] %s\n", i+1, instance)
	}
	fmt.Println("\nSelect instances to add (space-separated numbers, or type 'all' for all):")
	fmt.Print("> ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "all" {
		return instances
	}

	// ✅ Handle numbered selections
	selectedInstances := []string{}
	indices := strings.Fields(input)
	for _, index := range indices {
		if i, err := strconv.Atoi(index); err == nil && i > 0 && i <= len(instances) {
			selectedInstances = append(selectedInstances, instances[i-1])
		}
	}
	return selectedInstances
}
