package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bxtal-lsn/gosible/internal/executor"
	"github.com/bxtal-lsn/gosible/internal/inventory"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Ansible playbooks with optional auto-discovery and dry-run mode",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)
		var inventoryFile string
		var instances []string
		var playbooks []string
		var extraVars []string
		dryRun := false

		// ✅ Ask if the user has an existing inventory file
		fmt.Println("\n📂 Do you already have an inventory file? (yes/no)")
		fmt.Print("> ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "yes" {
			// ✅ Use existing inventory file
			fmt.Println("\n📍 Enter the path to your inventory file:")
			fmt.Print("> ")
			inventoryFile, _ = reader.ReadString('\n')
			inventoryFile = strings.TrimSpace(inventoryFile)
		} else {
			// ✅ No inventory file → Ask if user wants to auto-discover instances
			fmt.Println("\n🔍 Do you want to auto-discover running Multipass/Docker instances? (yes/no)")
			fmt.Print("> ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "yes" {
				instances = discoverInstances()
			} else {
				// ✅ Manually enter instances
				fmt.Println("\n🖥️ Enter server IPs or DNS names (space-separated):")
				fmt.Print("> ")
				input, _ := reader.ReadString('\n')
				instances = strings.Fields(strings.TrimSpace(input))
			}

			// ✅ Proceed with inventory creation
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
		}

		// ✅ Ask for playbooks
		fmt.Println("\n📜 Enter playbooks to run (space-separated):")
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		playbooks = strings.Fields(strings.TrimSpace(input))

		// ✅ Ask if user wants to enable dry-run mode
		fmt.Println("\n🔍 Would you like to run this in dry-run mode? (yes/no)")
		fmt.Print("> ")
		response, _ = reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "yes" {
			dryRun = true
			fmt.Println("✅ Dry-run mode enabled! Playbooks will simulate changes without applying them.")
		}

		// ✅ Execute playbooks
		for _, playbook := range playbooks {
			fmt.Printf("\n🚀 Running playbook: %s using inventory: %s\n", playbook, inventoryFile)
			executor.ExecuteAnsiblePlaybook(inventoryFile, playbook, extraVars, dryRun)
		}
	},
}

// ✅ Auto-discover Multipass/Docker instances
func discoverInstances() []string {
	var instances []string

	// ✅ Detect Multipass instances
	fmt.Println("\n🔍 Checking for running Multipass instances...")
	out, err := exec.Command("multipass", "list", "--format", "csv").Output()
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
	out, err = exec.Command("docker", "ps", "--format", "{{.Names}}").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if len(line) > 0 {
				instances = append(instances, strings.TrimSpace(line)) // Use container name
			}
		}
	}

	// ✅ Prompt user to select instances
	if len(instances) > 0 {
		fmt.Println("\n🔍 Found the following instances:")
		for i, instance := range instances {
			fmt.Printf("[%d] %s\n", i+1, instance)
		}
		fmt.Println("\nSelect instances to add (space-separated numbers, or type 'all' for all):")
		fmt.Print("> ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "all" {
			return instances
		}

		selectedInstances := []string{}
		indices := strings.Fields(input)
		for _, index := range indices {
			if i, err := strconv.Atoi(index); err == nil && i > 0 && i <= len(instances) {
				selectedInstances = append(selectedInstances, instances[i-1])
			}
		}
		return selectedInstances
	}

	fmt.Println("⚠️ No running Multipass or Docker instances found.")
	return []string{}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
