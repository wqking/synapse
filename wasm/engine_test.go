package wasm_test

import (
	"os"
	"testing"

	"github.com/phoreproject/synapse/wasm"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestEngine(t *testing.T) {
	engine := wasm.NewEngine()
	script, err := engine.LoadFromFile("/test.wasm")
	if err != nil {
		t.Fatal(err)
	}
	script.Run()
}
