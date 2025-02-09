package inventory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ✅ Test case: Create inventory file and check if it exists
func TestCreateInventoryFile(t *testing.T) {
	tempDir := t.TempDir()

	hosts := []HostConfig{
		{Host: "10.0.0.1", SSHUser: "ubuntu", SSHKeyFile: "~/.ssh/id_rsa", SSHPort: "22", Become: true},
		{Host: "10.0.0.2", SSHUser: "root", SSHKeyFile: "~/.ssh/id_rsa", Become: false},
		{Host: "10.0.0.3", Group: "web", SSHUser: "deployer", SSHKeyFile: "~/.ssh/id_rsa", SSHPort: "2222", Become: true},
	}

	// ✅ Create inventory file and handle error
	inventoryFile, err := CreateInventoryFile(tempDir, hosts)
	if err != nil {
		t.Fatalf("Error creating inventory file: %v", err)
	}

	// 🔍 Check if file exists
	if _, err := os.Stat(inventoryFile); os.IsNotExist(err) {
		t.Fatalf("Expected inventory file %s to be created, but it does not exist", inventoryFile)
	}

	// 🔍 Read file content
	content, err := os.ReadFile(inventoryFile)
	if err != nil {
		t.Fatalf("Error reading inventory file: %v", err)
	}

	// ✅ Ensure file contains expected host entries
	expectedEntries := []string{
		"10.0.0.1:",
		"ansible_user: ubuntu",
		"ansible_become: true",
		"10.0.0.2:",
		"ansible_user: root",
		"10.0.0.3:",
		"ansible_user: deployer",
		"ansible_port: 2222",
		"ansible_become: true",
		"web:",
	}
	for _, expected := range expectedEntries {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Inventory file missing expected entry: %s", expected)
		}
	}
}

// ✅ Test case: Ensure unique filenames are generated
func TestGetUniqueInventoryFilename(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, "inv.yml"), []byte{}, 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "inv1.yml"), []byte{}, 0o644)

	uniqueFile := getUniqueInventoryFilename(tempDir)
	expectedFile := filepath.Join(tempDir, "inv2.yml")
	if uniqueFile != expectedFile {
		t.Errorf("Expected unique filename %s, got %s", expectedFile, uniqueFile)
	}
}

// ✅ Test case: Handle empty hosts list gracefully
func TestCreateInventoryFile_EmptyHosts(t *testing.T) {
	tempDir := t.TempDir()

	hosts := []HostConfig{}

	// ✅ Create empty inventory file and handle error
	inventoryFile, err := CreateInventoryFile(tempDir, hosts)
	if err != nil {
		t.Fatalf("Error creating empty inventory file: %v", err)
	}

	// 🔍 Read file content
	content, err := os.ReadFile(inventoryFile)
	if err != nil {
		t.Fatalf("Error reading inventory file: %v", err)
	}

	// ✅ Ensure empty but valid YAML structure
	expectedStructure := "---\nall:\n  hosts:\n"
	if !strings.Contains(string(content), expectedStructure) {
		t.Errorf("Expected empty inventory structure, got:\n%s", content)
	}
}

// ✅ Test case: Handle duplicate hostnames
func TestCreateInventoryFile_DuplicateHosts(t *testing.T) {
	tempDir := t.TempDir()

	hosts := []HostConfig{
		{Host: "server1", SSHUser: "ubuntu"},
		{Host: "server1", SSHUser: "root"},
	}

	// ✅ Create inventory file and handle error
	inventoryFile, err := CreateInventoryFile(tempDir, hosts)
	if err != nil {
		t.Fatalf("Error creating inventory file: %v", err)
	}

	// 🔍 Read file content
	content, err := os.ReadFile(inventoryFile)
	if err != nil {
		t.Fatalf("Error reading inventory file: %v", err)
	}

	// ✅ Ensure both duplicates are present
	expectedEntries := []string{"server1:", "ansible_user: ubuntu", "ansible_user: root"}
	for _, expected := range expectedEntries {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Inventory file missing expected entry: %s", expected)
		}
	}
}

// ✅ Test case: Handle special characters in hostnames and groups
func TestCreateInventoryFile_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()

	hosts := []HostConfig{
		{Host: "web-server@company.com", SSHUser: "deploy"},
		{Host: "db-server_1", Group: "db_group!", SSHUser: "root"},
	}

	// ✅ Create inventory file and handle error
	inventoryFile, err := CreateInventoryFile(tempDir, hosts)
	if err != nil {
		t.Fatalf("Error creating inventory file: %v", err)
	}

	// 🔍 Read file content
	content, err := os.ReadFile(inventoryFile)
	if err != nil {
		t.Fatalf("Error reading inventory file: %v", err)
	}

	// ✅ Ensure the special characters are properly handled
	expectedEntries := []string{
		"web-server@company.com:",
		"ansible_user: deploy",
		"db-server_1:",
		"db_group!:",
	}
	for _, expected := range expectedEntries {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Inventory file missing expected entry: %s", expected)
		}
	}
}

// ✅ Test case: Handle very long hostnames and group names
func TestCreateInventoryFile_LongHostnames(t *testing.T) {
	tempDir := t.TempDir()

	longHostname := strings.Repeat("a", 256)
	longGroup := strings.Repeat("b", 256)

	hosts := []HostConfig{
		{Host: longHostname, Group: longGroup, SSHUser: "admin"},
	}

	// ✅ Create inventory file and handle error
	inventoryFile, err := CreateInventoryFile(tempDir, hosts)
	if err != nil {
		t.Fatalf("Error creating inventory file: %v", err)
	}

	// 🔍 Read file content
	content, err := os.ReadFile(inventoryFile)
	if err != nil {
		t.Fatalf("Error reading inventory file: %v", err)
	}

	// ✅ Ensure long names are correctly stored
	if !strings.Contains(string(content), longHostname) || !strings.Contains(string(content), longGroup) {
		t.Errorf("Expected long host/group names in inventory, but they are missing")
	}
}

// ✅ Test case: Handle invalid directory
func TestCreateInventoryFile_InvalidDirectory(t *testing.T) {
	invalidDir := "/root/protected"

	hosts := []HostConfig{
		{Host: "server1", SSHUser: "admin"},
	}

	// 🚨 Try to create inventory (should return an error)
	_, err := CreateInventoryFile(invalidDir, hosts)
	if err == nil {
		t.Errorf("Expected error when creating inventory in a protected directory, but got nil")
	}
}

