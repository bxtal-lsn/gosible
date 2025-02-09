package cmd

import (
	"bufio"
	"fmt"
	"os"
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

// ✅ Main logic for running playbooks
func runPlaybook(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)
	var inventoryFile string
	var instances []string
	dryRun := false

	// ✅ Ask for inventory file or create one
	inventoryFile = askForInventory(reader, &instances)

	// ✅ Ask for playbooks
	playbooks := askForPlaybooks(reader)

	// ✅ Ask if dry-run mode should be enabled
	dryRun = askForDryRun(reader)

	// ✅ Execute playbooks
	for _, playbook := range playbooks {
		fmt.Printf("\n🚀 Running playbook: %s using inventory: %s\n", playbook, inventoryFile)
		executor.ExecuteAnsiblePlaybook(inventoryFile, playbook, nil, dryRun)
	}
}

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
		*instances = inventory.DiscoverInstances()
	} else {
		fmt.Println("\n🖥️ Enter server IPs or DNS names (space-separated):")
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		*instances = strings.Fields(strings.TrimSpace(input))
	}

	// ✅ Proceed with inventory creation
	return createInventoryFile(reader, *instances)
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
