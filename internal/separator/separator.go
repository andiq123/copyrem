package separator

import (
	"context"
	"fmt"
)

const errDemucsNotFound = "demucs not found: pip install demucs, then ensure demucs or python -m demucs is on PATH"

func SeparateWithProgress(ctx context.Context, input, outVocals, outInstrumental string, onProgress func(int)) error {
	if !DemucsAvailable() {
		return fmt.Errorf("%s", errDemucsNotFound)
	}
	return SeparateWithDemucsPython(ctx, input, outVocals, outInstrumental, onProgress)
}
