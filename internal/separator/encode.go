package separator

import (
	"context"
	"fmt"
	"os/exec"

	"copyrem/internal/ffmpeg"
)

func encodeWavToMp3(ctx context.Context, wavPath, mp3Path string) error {
	bin := ffmpeg.FindBinary()
	cmd := exec.CommandContext(ctx, bin, "-y", "-i", wavPath, "-c:a", "libmp3lame", "-b:a", "192k", mp3Path)
	if out, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg encode: %w\n%s", err, out)
	}
	return nil
}
