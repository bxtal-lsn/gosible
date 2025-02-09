package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bxtal-lsn/gosible/internal/executor"
	"github.com/bxtal-lsn/gosible/internal/inventory"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Ansible playbooks with optional auto-discovery and dry-run mode",
	Run:   runPlaybook,
}

func runPlaybook(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)
	var inventoryFile string
	var instances []string
	var playbooks []string
	var dryRun bool

	// Check command history
	historyEntries, err := loadHistory()
	if err != nil {
		fmt.Printf("⚠️ Could not load command history: %v\n", err)
	}

	// Offer to reuse previous command if history exists
	if len(historyEntries) > 0 {
		fmt.Println("\n🕒 Previous commands (latest first):")
		displayedEntries := historyEntries
		if len(displayedEntries) > 5 {
			displayedEntries = displayedEntries[len(displayedEntries)-5:]
		}

		// Display entries in reverse chronological order
		for i := len(displayedEntries) - 1; i >= 0; i-- {
			entry := displayedEntries[i]
			fmt.Printf("%d. Inventory: %s | Playbooks: %s | Dry-run: %t\n",
				len(displayedEntries)-i,
				entry.InventoryFile,
				strings.Join(entry.Playbooks, " "),
				entry.DryRun)
		}

		fmt.Println("\n↩️ Choose a previous command (1-5) or press Enter to start fresh:")
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input != "" {
			if choice, err := strconv.Atoi(input); err == nil {
				if choice >= 1 && choice <= len(displayedEntries) {
					selectedIndex := len(displayedEntries) - choice
					selectedEntry := displayedEntries[selectedIndex]

					// Use selected history entry
					inventoryFile = selectedEntry.InventoryFile
					playbooks = selectedEntry.Playbooks
					dryRun = selectedEntry.DryRun

					// Execute directly
					for _, playbook := range playbooks {
						fmt.Printf("\n🚀 Running playbook: %s using inventory: %s\n",
							playbook, inventoryFile)
						executor.ExecuteAnsiblePlaybook(inventoryFile, playbook, nil, dryRun)
					}

					// Save to history again
					saveNewHistoryEntry(inventoryFile, playbooks, dryRun)
					return
				}
			}
		}
	}

	// Normal execution flow
	inventoryFile = askForInventory(reader, &instances)
	playbooks = askForPlaybooks(reader)
	dryRun = askForDryRun(reader)

	// Save to history
	saveNewHistoryEntry(inventoryFile, playbooks, dryRun)

	// Execute playbooks
	for _, playbook := range playbooks {
		fmt.Printf("\n🚀 Running playbook: %s using inventory: %s\n", playbook, inventoryFile)
		executor.ExecuteAnsiblePlaybook(inventoryFile, playbook, nil, dryRun)
		if dryRun {
			fmt.Println("\n🔄 Would you like to run this again without dry-run? (yes/no)")
			fmt.Print("> ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response == "yes" {
				// Re-run with same settings but dry-run disabled
				for _, playbook := range playbooks {
					fmt.Printf("\n🚀 Running playbook: %s using inventory: %s\n",
						playbook, inventoryFile)
					executor.ExecuteAnsiblePlaybook(inventoryFile, playbook, nil, false)
				}
				// Save new history entry for non-dry run
				saveNewHistoryEntry(inventoryFile, playbooks, false)
			}
		}

	}
}

// CommandHistoryEntry represents a previous command run
type CommandHistoryEntry struct {
	InventoryFile string   `json:"inventory_file"`
	Playbooks     []string `json:"playbooks"`
	DryRun        bool     `json:"dry_run"`
}

// getHistoryPath returns the path to the history file
func getHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gosible_history"), nil
}

// loadHistory loads up to the last 5 commands from the history file
func loadHistory() ([]CommandHistoryEntry, error) {
	path, err := getHistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []CommandHistoryEntry{}, nil
	} else if err != nil {
		return nil, err
	}

	var entries []CommandHistoryEntry
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry CommandHistoryEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid lines
		}
		entries = append(entries, entry)
	}

	// Keep only last 5 entries
	if len(entries) > 5 {
		entries = entries[len(entries)-5:]
	}

	return entries, nil
}

// saveHistory saves the command history to file
func saveHistory(entries []CommandHistoryEntry) error {
	path, err := getHistoryPath()
	if err != nil {
		return err
	}

	var lines []string
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		lines = append(lines, string(data))
	}

	data := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(data), 0o644)
}

// saveNewHistoryEntry adds a new entry to history
func saveNewHistoryEntry(inventoryFile string, playbooks []string, dryRun bool) {
	entry := CommandHistoryEntry{
		InventoryFile: inventoryFile,
		Playbooks:     playbooks,
		DryRun:        dryRun,
	}

	currentHistory, err := loadHistory()
	if err != nil {
		fmt.Printf("⚠️ Could not load command history: %v\n", err)
		currentHistory = []CommandHistoryEntry{}
	}

	currentHistory = append(currentHistory, entry)
	if len(currentHistory) > 5 {
		currentHistory = currentHistory[len(currentHistory)-5:]
	}

	if err := saveHistory(currentHistory); err != nil {
		fmt.Printf("⚠️ Could not save command history: %v\n", err)
	}
}

// ✅ Ask user for inventory file or create one
// ✅ Ask user for inventory file or create one
func askForInventory(reader *bufio.Reader, instances *[]string) string {
	fmt.Println("\n📂 Do you already have an inventory file? (yes/no)")
	fmt.Print("> ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "yes" {
		fmt.Println("\n📍 Enter the path to your inventory file:")
		fmt.Print("> ")
		inventoryFile, _ := reader.ReadString('\n')
		return strings.TrimSpace(inventoryFile)
	}

	// ✅ No inventory file → Ask if user wants to auto-discover instances
	fmt.Println("\n🔍 Do you want to auto-discover running Multipass/Docker instances? (yes/no)")
	fmt.Print("> ")
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "yes" {
		*instances = inventory.DiscoverInstances(reader) // ✅ Use `reader`
	} else {
		fmt.Println("\n🖥️ Enter server IPs or DNS names (space-separated):")
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		*instances = strings.Fields(strings.TrimSpace(input))
	}

	// ✅ Proceed with inventory creation
	return createInventoryFile(reader, *instances) // ✅ Use `reader`
}

// ✅ Create a new inventory file
func createInventoryFile(reader *bufio.Reader, instances []string) string {
	fmt.Println("\n📂 Where should the inventory file be saved? (Press Enter for current directory):")
	fmt.Print("> ")
	inventoryDir, _ := reader.ReadString('\n')
	inventoryDir = strings.TrimSpace(inventoryDir)
	if inventoryDir == "" {
		inventoryDir = "."
	}

	// ✅ Configure each instance
	hostConfigs := []inventory.HostConfig{}
	for _, instance := range instances {
		fmt.Printf("\n🖥️ Configuring %s\n", instance)

		fmt.Println("\n👤 SSH user (e.g., ubuntu, root):")
		fmt.Print("> ")
		sshUser, _ := reader.ReadString('\n')
		sshUser = strings.TrimSpace(sshUser)

		fmt.Println("\n🔑 SSH private key file (Press Enter for default ~/.ssh/id_rsa):")
		fmt.Print("> ")
		sshKey, _ := reader.ReadString('\n')
		sshKey = strings.TrimSpace(sshKey)
		if sshKey == "" {
			sshKey = "~/.ssh/id_rsa"
		}

		fmt.Println("\n📦 Server group (Press Enter to skip grouping):")
		fmt.Print("> ")
		group, _ := reader.ReadString('\n')
		group = strings.TrimSpace(group)

		fmt.Println("\n🔌 SSH port (Press Enter for default 22):")
		fmt.Print("> ")
		sshPort, _ := reader.ReadString('\n')
		sshPort = strings.TrimSpace(sshPort)

		fmt.Println("\n🔓 Enable sudo (become) for this server? (yes/no):")
		fmt.Print("> ")
		becomeInput, _ := reader.ReadString('\n')
		become := strings.TrimSpace(strings.ToLower(becomeInput)) == "yes"

		hostConfigs = append(hostConfigs, inventory.HostConfig{
			Host:       instance,
			Group:      group,
			SSHUser:    sshUser,
			SSHKeyFile: sshKey,
			SSHPort:    sshPort,
			Become:     become,
		})
	}

	// ✅ Create inventory file
	inventoryFile, err := inventory.CreateInventoryFile(inventoryDir, hostConfigs)
	if err != nil {
		fmt.Printf("❌ Error creating inventory file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Inventory file created at: %s\n", inventoryFile)
	return inventoryFile
}

// ✅ Ask user for playbooks to run
func askForPlaybooks(reader *bufio.Reader) []string {
	fmt.Println("\n📜 Enter playbooks to run (space-separated):")
	fmt.Print("> ")
	input, _ := reader.ReadString('\n')
	return strings.Fields(strings.TrimSpace(input))
}

// ✅ Ask if dry-run mode should be enabled
func askForDryRun(reader *bufio.Reader) bool {
	fmt.Println("\n🔍 Would you like to run this in dry-run mode? (yes/no)")
	fmt.Print("> ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response == "yes" {
		fmt.Println("✅ Dry-run mode enabled! Playbooks will simulate changes without applying them.")
		return true
	}
	return false
}

func init() {
	rootCmd.AddCommand(runCmd)
}
