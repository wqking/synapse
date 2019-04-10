package wasm

import (
	"fmt"
	"sync"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/sirupsen/logrus"
)

/*
objectMapper is used to pass immutable object from host to WASM
The host calls registerObject to get an ID, then pass the ID to WASM
WASM can pass the ID around, and pass back to the host
The host can call getObject to retrieve the object.
*/
type objectMapper struct {
	m    map[int64]interface{}
	lock *sync.Mutex
	id   int64
}

func newObjectMapper() *objectMapper {
	om := objectMapper{
		m:    map[int64]interface{}{},
		lock: new(sync.Mutex),
		id:   0,
	}
	return &om
}

func (om *objectMapper) registerObject(obj interface{}) int64 {
	om.lock.Lock()
	defer om.lock.Unlock()
	om.id++
	om.m[om.id] = obj
	return om.id
}

func (om *objectMapper) getObject(id int64) interface{} {
	obj, ok := om.m[id]
	if ok {
		return obj
	}

	return nil
}

func (om *objectMapper) freeObject(id int64) {
	_, ok := om.m[id]
	if ok {
		delete(om.m, id)
	} else {
		logrus.Warnf("Freeing unknown object ID: %d", id)
	}
}

// Script represents an instance of WASM script
type Script struct {
	module        *wasm.Module
	vm            *exec.VM
	objectMapper  *objectMapper
	memoryManager *MemoryManager
	// the per-script imported functions
	moduleImportMap *moduleImportMap
}

const globalModuleName = "env"

// NewScript creates a new script
func NewScript() *Script {
	script := &Script{
		moduleImportMap: newModuleImportMap(),
	}
	script.objectMapper = newObjectMapper()
	script.memoryManager = newMemoryManager(func() []byte {
		return script.vm.Memory()
	}, 32*1024)

	script.moduleImportMap.registerFunction(globalModuleName, FuncRegister{
		Name: "freeObject",
		F: func(process *exec.Process, id int64) {
			script.objectMapper.freeObject(id)
		},
	})
	script.moduleImportMap.registerFunction(globalModuleName, FuncRegister{
		Name: "allocate",
		F: func(process *exec.Process, size int32) int32 {
			return int32(script.memoryManager.Allocate(int(size)))
		},
	})
	script.moduleImportMap.registerFunction(globalModuleName, FuncRegister{
		Name: "free",
		F: func(process *exec.Process, addr int32) {
			script.memoryManager.Free(int(addr))
		},
	})

	return script
}

func (script *Script) init(module *wasm.Module, vm *exec.VM) {
	script.module = module
	script.vm = vm
	script.memoryManager.init()
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
