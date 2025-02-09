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
	Short: "Run Ansible playbooks on selected instances",
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)
		var inventoryFile string
		var instances []string
		var playbooks []string

		// ✅ Ask if the user has an inventory file or wants to create one
		fmt.Println("\n📂 Do you already have an inventory file? (yes/no)")
		fmt.Print("> ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "yes" {
			fmt.Println("\n📍 Enter the path to your inventory file:")
			fmt.Print("> ")
			inventoryFile, _ = reader.ReadString('\n')
			inventoryFile = strings.TrimSpace(inventoryFile)
		} else {
			fmt.Println("\n🖥️ Enter server IPs or DNS names (space-separated):")
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			instances = strings.Fields(strings.TrimSpace(input))

			// ✅ Ask where to save the inventory
			fmt.Println("\n📂 Where should the inventory file be saved? (Press Enter for current directory):")
			fmt.Print("> ")
			inventoryDir, _ := reader.ReadString('\n')
			inventoryDir = strings.TrimSpace(inventoryDir)
			if inventoryDir == "" {
				inventoryDir = "."
			}

			// ✅ Define hostConfigs before the loop
			hostConfigs := []inventory.HostConfig{}

			// ✅ Ask per-host configuration
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

			// ✅ Create inventory file with default `inventory.yml` but increment if necessary
			inventoryFile = inventory.CreateInventoryFile(inventoryDir, hostConfigs)
			fmt.Printf("\n✅ Inventory file created at: %s (Default: inventory.yml, increments if necessary)\n", inventoryFile)
		}

		// ✅ Ask for playbooks
		fmt.Println("\n📜 Enter playbooks to run (space-separated):")
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		playbooks = strings.Fields(strings.TrimSpace(input))

		// ✅ Execute playbooks
		for _, playbook := range playbooks {
			fmt.Printf("\n🚀 Running playbook: %s using inventory: %s\n", playbook, inventoryFile)
			executor.ExecuteAnsiblePlaybook(inventoryFile, playbook, nil)
		}
	},
}

