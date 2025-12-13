package ffmpeg

import (
	"fmt"
	"os"
)

func checkFFmpegBinary(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFFmpegNotFound
		}
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("ffmpeg path is a directory: %s", path)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("ffmpeg binary is not executable: %s", path)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
