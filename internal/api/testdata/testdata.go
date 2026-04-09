package testdata

import (
	"os"
	"path/filepath"
	"runtime"
)

// Load reads a fixture file from the testdata directory.
func Load(name string) []byte {
	_, thisFile, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(thisFile), name)
	data, err := os.ReadFile(path)
	if err != nil {
		panic("testdata.Load: " + err.Error())
	}
	return data
}
