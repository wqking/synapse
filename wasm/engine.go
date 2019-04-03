package wasm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/sirupsen/logrus"
)

// Script represents an instance of WASM script
type Script struct {
	module *wasm.Module
	vm     *exec.VM
}

// Run executes the script
func (script *Script) Run(args ...uint64) (interface{}, error) {
	return script.RunByName("Entry", args...)
}

// RunByName executes the script by the entryName
func (script *Script) RunByName(entryName string, args ...uint64) (interface{}, error) {
	for name, e := range script.module.Export.Entries {
		if name != entryName {
			continue
		}

		i := int64(e.Index)
		output, err := script.vm.ExecCode(i, args...)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Printf("%s: %v (%T)\n", name, output, output)

		return output, err
	}

	return nil, nil
}

// FuncRegister is the entry to register a function
type FuncRegister struct {
	Name        string
	F           interface{}
	ParamTypes  []wasm.ValueType
	ReturnTypes []wasm.ValueType
}

func (fr *FuncRegister) importer() (*wasm.Module, error) {
	m := wasm.NewModule()
	m.Types = &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{
			{
				Form:        0,
				ParamTypes:  fr.ParamTypes,
				ReturnTypes: fr.ReturnTypes,
			},
		},
	}
	m.FunctionIndexSpace = []wasm.Function{
		{
			Sig:  &m.Types.Entries[0],
			Host: reflect.ValueOf(fr.F),
			Body: &wasm.FunctionBody{},
		},
	}
	m.Export = &wasm.SectionExports{
		Entries: map[string]wasm.ExportEntry{
			"_native": {
				FieldStr: "_naive",
				Kind:     wasm.ExternalFunction,
				Index:    0,
			},
		},
	}

	return m, nil
}

func (fr *FuncRegister) solve() error {
	if fr.ParamTypes == nil {
		t := reflect.TypeOf(fr.F)
		fr.ParamTypes = []wasm.ValueType{}
		// Skip first parameter since it's exec.Process
		for i := 1; i < t.NumIn(); i++ {
			in := t.In(i)
			k, err := fr.goKindToWASM(in.Kind())
			if err != nil {
				return err
			}
			fr.ParamTypes = append(fr.ParamTypes, k)
		}
	}

	if fr.ReturnTypes == nil {
		t := reflect.TypeOf(fr.F)
		fr.ReturnTypes = []wasm.ValueType{}
		for i := 0; i < t.NumOut(); i++ {
			out := t.Out(i)
			k, err := fr.goKindToWASM(out.Kind())
			if err != nil {
				return err
			}
			fr.ReturnTypes = append(fr.ReturnTypes, k)
		}
	}

	return nil
}

func (fr *FuncRegister) goKindToWASM(k reflect.Kind) (wasm.ValueType, error) {
	switch k {
	case reflect.Bool, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32:
		return wasm.ValueTypeI32, nil

	// assume 64 bits compiler
	case reflect.Int, reflect.Int64, reflect.Uint64:
		return wasm.ValueTypeI64, nil

	case reflect.Float32:
		return wasm.ValueTypeF32, nil

	case reflect.Float64:
		return wasm.ValueTypeF64, nil
	}

	return wasm.ValueTypeI32, fmt.Errorf("Unknow type: %v", k)
}

type moduleImport struct {
	functionImportMap map[string]FuncRegister
}

func (module *moduleImport) importer() (*wasm.Module, error) {
	m := wasm.NewModule()
	m.Types = &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{},
	}
	m.FunctionIndexSpace = []wasm.Function{}
	m.Export = &wasm.SectionExports{
		Entries: map[string]wasm.ExportEntry{},
	}

	index := 0
	for name, fr := range module.functionImportMap {
		m.Types.Entries = append(m.Types.Entries, wasm.FunctionSig{
			Form:        0,
			ParamTypes:  fr.ParamTypes,
			ReturnTypes: fr.ReturnTypes,
		})
		m.FunctionIndexSpace = append(m.FunctionIndexSpace, wasm.Function{
			Sig:  &m.Types.Entries[index],
			Host: reflect.ValueOf(fr.F),
			Body: &wasm.FunctionBody{},
		})
		m.Export.Entries[name] = wasm.ExportEntry{
			FieldStr: name,
			Kind:     wasm.ExternalFunction,
			Index:    uint32(index),
		}

		index++
	}

	return m, nil
}

// Engine is the WASM engine
type Engine struct {
	moduleImportMap map[string]moduleImport
}

// NewEngine creates an Engine
func NewEngine() *Engine {
	return &Engine{
		moduleImportMap: map[string]moduleImport{},
	}
}

func (engine *Engine) importer(name string) (*wasm.Module, error) {
	fmt.Println("WASM engine, import:  " + name)

	module, ok := engine.moduleImportMap[name]
	if ok {
		return module.importer()
	}
	return nil, fmt.Errorf("Unkown import symbol %s", name)
}

// LoadFromFile loads script from .wasm file
func (engine *Engine) LoadFromFile(fileName string) (*Script, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return engine.LoadFromReader(f)
}

// LoadFromBytes loads script from bytes buffer
func (engine *Engine) LoadFromBytes(buffer []byte) (*Script, error) {
	reader := bytes.NewReader(buffer)
	return engine.LoadFromReader(reader)
}

// LoadFromReader loads script from a reader
func (engine *Engine) LoadFromReader(reader io.Reader) (*Script, error) {
	m, err := wasm.ReadModule(reader, engine.importer)
	if err != nil {
		return nil, err
	}

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

// RegisterFunction registers a function to import
func (engine *Engine) RegisterFunction(moduleName string, entry FuncRegister) error {
	_, ok := engine.moduleImportMap[moduleName]
	if !ok {
		engine.moduleImportMap[moduleName] = moduleImport{
			functionImportMap: make(map[string]FuncRegister),
		}
	}
	err := entry.solve()
	if err != nil {
		logrus.Errorf("RegisterFunction: error %v", err)
		return err
	}
	engine.moduleImportMap[moduleName].functionImportMap[entry.Name] = entry
	return nil
}

// RegisterFunctions registers functions to import
func (engine *Engine) RegisterFunctions(moduleName string, entries []FuncRegister) error {
	_, ok := engine.moduleImportMap[moduleName]
	if !ok {
		engine.moduleImportMap[moduleName] = moduleImport{
			functionImportMap: make(map[string]FuncRegister),
		}
	}

	for _, entry := range entries {
		err := entry.solve()
		if err != nil {
			logrus.Errorf("RegisterFunction: error %v", err)
			return err
		}
		engine.moduleImportMap[moduleName].functionImportMap[entry.Name] = entry
	}

	return nil
}
