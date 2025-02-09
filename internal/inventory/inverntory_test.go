package inventory

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ✅ Properly mock `execCommand`
func mockExecCommand(name string, arg ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, arg...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

// ✅ Simulated process for mocking `exec.Command`
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	switch os.Args[3] {
	case "multipass":
		os.Stdout.Write([]byte("Name,State,IPv4\ninstance1,Running,10.0.0.5\ninstance2,Running,10.0.0.6\n"))
	case "docker":
		os.Stdout.Write([]byte("container1\ncontainer2\n"))
	}
	os.Exit(0)
}

// ✅ Test `DiscoverInstances`
func TestDiscoverInstances(t *testing.T) {
	// ✅ Override `execCommand`
	oldExecCommand := execCommand
	execCommand = mockExecCommand
	defer func() { execCommand = oldExecCommand }()

	// ✅ Simulate user input selecting "all"
	reader := bufio.NewReader(strings.NewReader("all\n"))

	// ✅ Pass the reader to `DiscoverInstances`
	instances := DiscoverInstances(reader)

	// ✅ Check that instances match expected mock data
	expectedInstances := []string{"10.0.0.5", "10.0.0.6", "container1", "container2"}
	for _, expected := range expectedInstances {
		found := false
		for _, instance := range instances {
			if instance == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected instance %s not found in discovered instances: %v", expected, instances)
		}
	}
}
