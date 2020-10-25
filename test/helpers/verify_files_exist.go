package helpers

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
)

func VerifyFilesExist(t TestingT, directories ...string) bool {
	require.NotZero(t, len(directories), "number of volumes should be more than 0")

	var filenames []string

	for _, dir := range directories {
		_, err := os.Lstat(dir)
		if err != nil {

		}

		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				filenames = append(filenames, path)
			}

			return err
		}); err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				require.NoError(t, err)
			}
		}
	}

	fmt.Println("files found:", filenames)

	return len(filenames) != 0
}
