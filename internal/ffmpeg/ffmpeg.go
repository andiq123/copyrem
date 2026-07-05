package ffmpeg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	pathOnce   sync.Once
	ffmpegBin  string
	ffprobeBin string
)

func initPaths() {
	pathOnce.Do(func() {
		ffmpegBin = find("ffmpeg")
		ffprobeBin = find("ffprobe")
		if _, err := exec.LookPath("ffprobe"); err != nil {
			if p := filepath.Join(filepath.Dir(ffmpegBin), "ffprobe"); fileExists(p) {
				ffprobeBin = p
			}
		}
	})
}

func FindBinary() string {
	initPaths()
	return ffmpegBin
}

func Duration(_ string, path string) (time.Duration, error) {
	initPaths()
	var buf bytes.Buffer
	cmd := exec.Command(ffprobeBin, "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", path)
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

func find(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	for _, dir := range searchDirs() {
		p := filepath.Join(dir, name)
		if fileExists(p) {
			return p
		}
	}
	return name
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func searchDirs() []string {
	dirs := []string{}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "bin"))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "bin"))
	}
	return dirs
}
