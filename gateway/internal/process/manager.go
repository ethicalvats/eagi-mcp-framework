package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

// DomainProcess represents a running Node.js MCP server
type DomainProcess struct {
	Name      string
	Cmd       *exec.Cmd
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser
	Restart   bool
	mu        sync.Mutex
	cancel    context.CancelFunc
	ctx       context.Context
}

type Manager struct {
	processes map[string]*DomainProcess
	mu        sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		processes: make(map[string]*DomainProcess),
	}
}

// StartDomain launches a Node.js process using `npx eagi serve --domain <name>`
func (m *Manager) StartDomain(name string, projectDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.processes[name]; exists {
		return fmt.Errorf("domain %s is already running", name)
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	// Assuming eagi is installed locally or via npx
	cmd := exec.CommandContext(ctx, "npx", "eagi", "serve", "--domain", name)
	cmd.Dir = projectDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	dp := &DomainProcess{
		Name:    name,
		Cmd:     cmd,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Restart: true,
		ctx:     ctx,
		cancel:  cancel,
	}

	m.processes[name] = dp

	// Start monitoring the process
	go m.monitorProcess(dp, projectDir)

	log.Printf("[Process Manager] Started domain %s (PID: %d)\n", name, cmd.Process.Pid)
	return nil
}

func (m *Manager) monitorProcess(dp *DomainProcess, projectDir string) {
	// Log stderr
	go func() {
		scanner := bufio.NewScanner(dp.Stderr)
		for scanner.Scan() {
			log.Printf("[%s STDERR] %s\n", dp.Name, scanner.Text())
		}
	}()

	err := dp.Cmd.Wait()
	log.Printf("[Process Manager] Domain %s exited: %v\n", dp.Name, err)

	m.mu.Lock()
	if dp.Restart {
		log.Printf("[Process Manager] Restarting domain %s in 2 seconds...\n", dp.Name)
		delete(m.processes, dp.Name)
		m.mu.Unlock()
		
		time.Sleep(2 * time.Second)
		m.StartDomain(dp.Name, projectDir)
	} else {
		delete(m.processes, dp.Name)
		m.mu.Unlock()
	}
}

// StopDomain gracefully shuts down a domain process
func (m *Manager) StopDomain(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if dp, exists := m.processes[name]; exists {
		dp.mu.Lock()
		dp.Restart = false
		dp.mu.Unlock()
		
		dp.cancel()
		log.Printf("[Process Manager] Stopped domain %s\n", name)
	}
}

func (m *Manager) GetProcess(name string) (*DomainProcess, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if dp, exists := m.processes[name]; exists {
		return dp, nil
	}
	return nil, fmt.Errorf("domain %s not running", name)
}
