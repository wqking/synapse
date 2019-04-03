package wasm_test

import (
	"os"
)

func TestMain(m *testing.M) {
	retCode := m.Run()
	os.Exit(retCode)
}

