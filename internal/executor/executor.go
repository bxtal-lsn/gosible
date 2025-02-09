package executor

import (
	"fmt"
	"os"
	"os/exec"
)

func ExecuteAnsiblePlaybook(inventory, playbook string, vars []string) {
	cmdArgs := []string{"-i", inventory, playbook}
	for _, v := range vars {
		cmdArgs = append(cmdArgs, "--extra-vars", v)
	}

	cmd := exec.Command("ansible-playbook", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing playbook:", err)
	}
}
