package ttyd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/es6kr/pmux/internal/session"
)

const (
	pidDir = "/tmp/pmux-ttyd"
)

// Instance represents a running ttyd instance
type Instance struct {
	Channel string `json:"channel"`
	Port    int    `json:"port"`
	PID     int    `json:"pid"`
	Socket  string `json:"socket"`
}

func init() {
	os.MkdirAll(pidDir, 0755)
}

func pidFile(channel string) string {
	return filepath.Join(pidDir, channel+".json")
}

// Start starts a ttyd instance for a session
func Start(channel string, port int) error {
	socket := session.GetSocketPath(channel)

	// Check if session exists
	if _, err := os.Stat(socket); os.IsNotExist(err) {
		return fmt.Errorf("session '%s' does not exist, create it first with: pmux start %s", channel, channel)
	}

	// Check if ttyd is already running for this channel
	if inst, _ := GetInstance(channel); inst != nil {
		return fmt.Errorf("ttyd already running for '%s' on port %d (PID: %d)", channel, inst.Port, inst.PID)
	}

	// Start ttyd process
	// ttyd -S <socket> tmux -S <socket> attach -t <channel>
	cmd := exec.Command("ttyd",
		"-p", strconv.Itoa(port),
		"-t", "fontSize=14",
		"-t", "theme={\"background\":\"#1a1b26\"}",
		"tmux", "-S", socket, "attach", "-t", channel,
	)

	// Detach from parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ttyd: %w", err)
	}

	// Save instance info
	inst := Instance{
		Channel: channel,
		Port:    port,
		PID:     cmd.Process.Pid,
		Socket:  socket,
	}

	data, _ := json.Marshal(inst)
	if err := os.WriteFile(pidFile(channel), data, 0644); err != nil {
		return fmt.Errorf("failed to save pid file: %w", err)
	}

	fmt.Printf("ttyd started for '%s' at http://localhost:%d (PID: %d)\n", channel, port, inst.PID)
	return nil
}

// Stop stops a ttyd instance
func Stop(channel string) error {
	inst, err := GetInstance(channel)
	if err != nil {
		return err
	}
	if inst == nil {
		return fmt.Errorf("no ttyd running for '%s'", channel)
	}

	// Kill the process
	process, err := os.FindProcess(inst.PID)
	if err != nil {
		os.Remove(pidFile(channel))
		return nil
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		os.Remove(pidFile(channel))
		return nil
	}

	os.Remove(pidFile(channel))
	fmt.Printf("ttyd stopped for '%s' (was PID: %d)\n", channel, inst.PID)
	return nil
}

// GetInstance returns the running instance for a channel
func GetInstance(channel string) (*Instance, error) {
	data, err := os.ReadFile(pidFile(channel))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var inst Instance
	if err := json.Unmarshal(data, &inst); err != nil {
		return nil, err
	}

	// Check if process is still running
	process, err := os.FindProcess(inst.PID)
	if err != nil {
		os.Remove(pidFile(channel))
		return nil, nil
	}

	// Check if process is alive (signal 0)
	if err := process.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidFile(channel))
		return nil, nil
	}

	return &inst, nil
}

// List lists all running ttyd instances
func List() {
	files, err := filepath.Glob(filepath.Join(pidDir, "*.json"))
	if err != nil || len(files) == 0 {
		fmt.Println("No ttyd instances running")
		return
	}

	fmt.Printf("%-15s %-10s %-10s %-30s\n", "CHANNEL", "PORT", "PID", "URL")

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var inst Instance
		if err := json.Unmarshal(data, &inst); err != nil {
			continue
		}

		// Verify process is still running
		if _, err := GetInstance(inst.Channel); err != nil || inst.PID == 0 {
			continue
		}

		fmt.Printf("%-15s %-10d %-10d http://localhost:%d\n",
			inst.Channel, inst.Port, inst.PID, inst.Port)
	}
}

// StopAll stops all ttyd instances
func StopAll() {
	files, _ := filepath.Glob(filepath.Join(pidDir, "*.json"))
	for _, f := range files {
		channel := filepath.Base(f)
		channel = channel[:len(channel)-5] // remove .json
		Stop(channel)
	}
}
