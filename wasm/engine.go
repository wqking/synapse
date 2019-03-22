package wasm

import (
	"fmt"
	"io/ioutil"

	"github.com/perlin-network/life/exec"
)

// Script represents an instance of WASM script
type Script struct {
	vm *exec.VirtualMachine
}

// Run executes the script
func (script *Script) Run() error {
	id, ok := vm.GetFunctionExport("main")
	if !ok {
		return errors.New("Can't find 'main' in WASM")
	}
	script.vm.Run(id)

	return nil
}

// Engine is the WASM engine
type Engine struct {
}

// NewEngine creates an Engine
func NewEngine() *Engine {
	return Engine {}
}

// LoadFromFile loads script from .wasm file
func (engine *Engine) LoadFromFile(fileName string) (*Script, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return engine.LoadFromBytes(bytes)
}

// LoadFromBytes loads script from bytes buffer
func (engine *Engine) LoadFromBytes(buffer []byte) (*Script, error) {
	vm, err := exec.NewVirtualMachine(bytes, exec.VMConfig{}, new(exec.NopResolver))
	if err != nil {
		return nil, err
	}

	script := &Script {
		vm: vm
	}

	return script, nil
}
