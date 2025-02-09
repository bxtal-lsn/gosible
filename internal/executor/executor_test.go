package executor

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ✅ Mock function to replace exec.Command
func mockExecCommand(name string, arg ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, arg...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

// ✅ Helper process to simulate command execution
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

// ✅ Capture stdout output
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run function
	f()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// ✅ Test execution of a playbook without extra-vars
func TestExecuteAnsiblePlaybook_Basic(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	// Capture the output
	output := captureOutput(func() {
		ExecuteAnsiblePlaybook("test_inventory.yml", "test_playbook.yml", nil, false)
	})

	expected := "🔄 Executing: ansible-playbook -i test_inventory.yml test_playbook.yml"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// ✅ Test execution with extra-vars
func TestExecuteAnsiblePlaybook_ExtraVars(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	output := captureOutput(func() {
		ExecuteAnsiblePlaybook("test_inventory.yml", "test_playbook.yml", []string{"key1=value1", "key2=value2"}, false)
	})

	expected := "🔄 Executing: ansible-playbook -i test_inventory.yml test_playbook.yml --extra-vars key1=value1 --extra-vars key2=value2"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// ✅ Test execution in dry-run mode
func TestExecuteAnsiblePlaybook_DryRun(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	output := captureOutput(func() {
		ExecuteAnsiblePlaybook("test_inventory.yml", "test_playbook.yml", nil, true)
	})

	expected := "🔄 Executing: ansible-playbook -i test_inventory.yml test_playbook.yml --check"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

