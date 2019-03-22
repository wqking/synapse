package wasm

import (
	"errors"
	"os"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
)

// Script represents an instance of WASM script
type Script struct {
	module *wasm.Module
	vm     *exec.VM
}

// Run executes the script
func (script *Script) Run() error {
	for _, e := range script.module.Export.Entries {
		i := int64(e.Index)
		script.vm.ExecCode(i)
		/*
			fidx := m.Function.Types[int(i)]
			ftype := m.Types.Entries[int(fidx)]
			switch len(ftype.ReturnTypes) {
			case 1:
				fmt.Fprintf(w, "%s() %s => ", name, ftype.ReturnTypes[0])
			case 0:
				fmt.Fprintf(w, "%s() => ", name)
			default:
				log.Printf("running exported functions with more than one return value is not supported")
				continue
			}
			if len(ftype.ParamTypes) > 0 {
				log.Printf("running exported functions with input parameters is not supported")
				continue
			}
			o, err := vm.ExecCode(i)
			if err != nil {
				fmt.Fprintf(w, "\n")
				log.Printf("err=%v", err)
				continue
			}
			if len(ftype.ReturnTypes) == 0 {
				fmt.Fprintf(w, "\n")
				continue
			}
			fmt.Fprintf(w, "%[1]v (%[1]T)\n", o)
		*/
	}

	return nil
}

// Engine is the WASM engine
type Engine struct {
}

// NewEngine creates an Engine
func NewEngine() *Engine {
	return &Engine{}
}

func importer(name string) (*wasm.Module, error) {
	f, err := os.Open(name + ".wasm")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, err := wasm.ReadModule(f, nil)
	if err != nil {
		return nil, err
	}
	err = validate.VerifyModule(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// LoadFromFile loads script from .wasm file
func (engine *Engine) LoadFromFile(fileName string) (*Script, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := wasm.ReadModule(f, importer)
	if err != nil {
		return nil, err
	}

	/*
		if verify {
			err = validate.VerifyModule(m)
			if err != nil {
				return nil, err
			}
		}
	*/

	if m.Export == nil {
		return nil, errors.New("module has no export section")
	}

	vm, err := exec.NewVM(m)
	if err != nil {
		return nil, err
	}

	script := &Script{
		module: m,
		vm:     vm,
	}

	return script, nil
}

// LoadFromBytes loads script from bytes buffer
func (engine *Engine) LoadFromBytes(buffer []byte) (*Script, error) {
	return nil, nil
}
