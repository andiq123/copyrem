package separator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	demucsVocalsFile     = "vocals.wav"
	demucsNoVocalsFile   = "no_vocals.wav"
	demucsProgressCap    = 70
	demucsProgressWindow = 5 * time.Minute
	demucsProgressTick   = 2 * time.Second
)

func findDemucsCmd() (name string, args []string) {
	if path, err := exec.LookPath("demucs"); err == nil {
		return path, nil
	}
	for _, py := range []string{"python3", "python"} {
		if path, err := exec.LookPath(py); err == nil {
			return path, []string{"-m", "demucs"}
		}
	}
	return "", nil
}

func DemucsAvailable() bool {
	name, _ := findDemucsCmd()
	return name != ""
}

func reportProgress(onProgress func(int), pct int) {
	if onProgress != nil {
		onProgress(pct)
	}
}

func findStemPaths(outDir string) (vocals, noVocals string, err error) {
	var foundVocals, foundNoVocals string
	err = filepath.Walk(outDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		switch filepath.Base(path) {
		case demucsVocalsFile:
			foundVocals = path
		case demucsNoVocalsFile:
			foundNoVocals = path
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}
	if foundVocals == "" || foundNoVocals == "" {
		return "", "", fmt.Errorf("demucs output missing: expected %s and %s under %s", demucsVocalsFile, demucsNoVocalsFile, outDir)
	}
	return foundVocals, foundNoVocals, nil
}

func SeparateWithDemucsPython(ctx context.Context, input, outVocals, outInstrumental string, onProgress func(int)) error {
	name, cmdArgs := findDemucsCmd()
	if name == "" {
		return fmt.Errorf("%s", errDemucsNotFound)
	}

	dir := filepath.Dir(input)
	base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	outDir := filepath.Join(dir, base+"_demucs_out")
	defer os.RemoveAll(outDir)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("demucs out dir: %w", err)
	}

	reportProgress(onProgress, 5)
	args := append(cmdArgs, "--two-stems=vocals", "-o", outDir, input)
	cmd := exec.CommandContext(ctx, name, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("demucs start: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ticker := time.NewTicker(demucsProgressTick)
	defer ticker.Stop()
	start := time.Now()
	var demucsDone bool
	for !demucsDone {
		select {
		case err := <-done:
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("demucs: %w", err)
			}
			reportProgress(onProgress, demucsProgressCap)
			demucsDone = true
		case <-ticker.C:
			elapsed := time.Since(start)
			pct := 10 + int(60*elapsed.Seconds()/demucsProgressWindow.Seconds())
			if pct > demucsProgressCap {
				pct = demucsProgressCap
			}
			reportProgress(onProgress, pct)
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			<-done
			return ctx.Err()
		}
	}

	vocalsWav, noVocalsWav, err := findStemPaths(outDir)
	if err != nil {
		return err
	}
	reportProgress(onProgress, 75)
	if err := encodeWavToMp3(ctx, vocalsWav, outVocals); err != nil {
		return err
	}
	reportProgress(onProgress, 87)
	if err := encodeWavToMp3(ctx, noVocalsWav, outInstrumental); err != nil {
		return err
	}
	reportProgress(onProgress, 100)
	return nil
}
