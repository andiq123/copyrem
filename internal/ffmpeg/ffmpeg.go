package ffmpeg

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func FindBinary() string {
	if p := nextToExecutable(); p != "" {
		return p
	}
	if p := nextToCwd(); p != "" {
		return p
	}
	return "ffmpeg"
}

func nextToExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(filepath.Dir(exe), "bin", "ffmpeg")
	if _, err := os.Stat(candidate); err != nil {
		return ""
	}
	return candidate
}

func nextToCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(cwd, "bin", "ffmpeg")
	if _, err := os.Stat(candidate); err != nil {
		return ""
	}
	return candidate
}

func Run(binary string, args []string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	cmd := exec.Command(binary, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
