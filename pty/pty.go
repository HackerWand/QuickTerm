package pty

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/creack/pty"
)

type PTY struct {
	ptmx *os.File
	cmd  *exec.Cmd
}

func New(shell string, cwd string, env []string) (*PTY, error) {
	if shell == "" {
		shell = getDefaultShell()
	}

	cmd := exec.Command(shell, "-l")
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), env...)
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &PTY{
		ptmx: ptmx,
		cmd:  cmd,
	}, nil
}

func (p *PTY) Read(b []byte) (int, error) {
	return p.ptmx.Read(b)
}

func (p *PTY) Write(b []byte) (int, error) {
	return p.ptmx.Write(b)
}

func (p *PTY) Resize(rows, cols int) error {
	return pty.Setsize(p.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func (p *PTY) Close() error {
	if p.ptmx != nil {
		p.ptmx.Close()
	}
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Signal(syscall.SIGTERM)
		p.cmd.Wait()
	}
	return nil
}

func getDefaultShell() string {
	switch runtime.GOOS {
	case "windows":
		return "powershell.exe"
	case "darwin":
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/zsh"
		}
		return shell
	default:
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
		return shell
	}
}
