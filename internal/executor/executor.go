package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ✅ Execute Ansible playbook, supporting dry-run mode
func ExecuteAnsiblePlaybook(inventory string, playbook string, vars []string, dryRun bool) {
	cmdArgs := []string{"-i", inventory, playbook}

	// ✅ Add extra variables
	for _, v := range vars {
		cmdArgs = append(cmdArgs, "--extra-vars", v)
	}

	// ✅ Enable dry-run mode if selected
	if dryRun {
		cmdArgs = append(cmdArgs, "--check")
	}

	cmd := exec.Command("ansible-playbook", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🔄 Executing: ansible-playbook %s\n", strings.Join(cmdArgs, " "))

	// ✅ Run command
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Error executing playbook:", err)
	}
}

