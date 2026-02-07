package ffmpeg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func FindBinary() string {
	return findBin("ffmpeg")
}

func Duration(binary, path string) (time.Duration, error) {
	probe := siblingBin(binary, "ffprobe")
	var buf bytes.Buffer
	cmd := exec.Command(probe, "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", path)
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}
	secs, err := strconv.ParseFloat(strings.TrimSpace(buf.String()), 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe: invalid duration %q", buf.String())
	}
	return time.Duration(secs * float64(time.Second)), nil
}

func findBin(name string) string {
	for _, dir := range searchDirs() {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return name
}

func siblingBin(binary, name string) string {
	candidate := filepath.Join(filepath.Dir(binary), name)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return name
}

func searchDirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "bin"))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "bin"))
	}
	return dirs
}
