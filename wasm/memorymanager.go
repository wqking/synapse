package wasm

import (
	"encoding/binary"

	"github.com/go-interpreter/wagon/exec"
)

var endianess = binary.LittleEndian

type memoryGetter func() []byte

type memoryChunk struct {
	size         int
	previousSize int
	free         bool
}

// MemoryManager manages heap for both host and WASM
type MemoryManager struct {
	getter memoryGetter
	// The memory within [0, top) is reserved for WASM internal usage
	// MemoryManager manages [top, memory size)
	top int

	chunks map[int]memoryChunk
}

// newMemoryManager creates a MemoryManager
func newMemoryManager(getter memoryGetter, top int) *MemoryManager {
	mm := &MemoryManager{
		getter: getter,
		top:    top,
	}
	return mm
}

// NewMemoryManagerForVM creates a MemoryManager
func NewMemoryManagerForVM(vm *exec.VM, top int) *MemoryManager {
	mm := &MemoryManager{
		getter: func() []byte {
			return vm.Memory()
		},
		top: top,
	}
	mm.init()
	return mm
}

// NewMemoryManagerForBuffer creates a MemoryManager
func NewMemoryManagerForBuffer(buffer []byte, top int) *MemoryManager {
	mm := &MemoryManager{
		getter: func() []byte {
			return buffer
		},
		top: top,
	}
	mm.init()
	return mm
}

func (mm *MemoryManager) init() {
	mm.chunks = map[int]memoryChunk{}

	mm.chunks[mm.top] = memoryChunk{
		len(mm.GetMemory()) - mm.top,
		0,
		true,
	}
}

// GetMemory gets the underlying memory
func (mm *MemoryManager) GetMemory() []byte {
	return mm.getter()
}

// Allocate allocates a block of memory
func (mm *MemoryManager) Allocate(size int) int {
	if size < 1 {
		size = 1
	}

	for addr, chunk := range mm.chunks {
		if chunk.free && chunk.size >= size {
			allocatedChunk := chunk
			allocatedChunk.size = size
			allocatedChunk.free = false
			mm.chunks[addr] = allocatedChunk

			if chunk.size > size {
				remainingChunk := memoryChunk{
					size:         chunk.size - size,
					previousSize: size,
					free:         true,
				}
				mm.chunks[addr+size] = remainingChunk
				mm.mergeFreeChunks(addr + size)
			}

			return addr
		}
	}

	return -1
}

// Free frees a block of memory
func (mm *MemoryManager) Free(addr int) {
	chunk, ok := mm.chunks[addr]
	if ok && !chunk.free {
		chunk.free = true
		mm.chunks[addr] = chunk
		mm.mergeFreeChunks(addr)
	}
}

func (mm *MemoryManager) mergeFreeChunks(addr int) {
	chunk := mm.chunks[addr]
	if !chunk.free {
		return
	}
	if chunk.previousSize > 0 {
		previousAddr := addr - chunk.previousSize
		previousChunk, ok := mm.chunks[previousAddr]
		if ok && previousChunk.free {
			chunk.previousSize = previousChunk.previousSize
			chunk.size = previousChunk.size + chunk.size
			delete(mm.chunks, addr)
			mm.chunks[previousAddr] = chunk
			addr = previousAddr
		}
	}
	if chunk.size > 0 {
		nextAddr := addr + chunk.size
		nextChunk, ok := mm.chunks[nextAddr]
		if ok && nextChunk.free {
			chunk.size = nextChunk.size + chunk.size
			delete(mm.chunks, nextAddr)
			mm.chunks[addr] = chunk
		}
	}
}
