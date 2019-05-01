package wasm

import (
	"math"
	"fmt"
	"github.com/phoreproject/synapse/chainhash"
)

// ReadBytes reads bytes
func ReadBytes(memory []byte, offset int, length int) []byte {
	memoryLen := len(memory)
	if length > memoryLen - offset {
		length = memoryLen - offset
	}
	if length <= 0 {
		return []byte{}
	}

	return memory[offset: offset+length]
}

// WriteBytes writes bytes
func WriteBytes(memory []byte, offset int, data []byte) {
	memoryLen := len(memory)
	length := len(data)
	if length > memoryLen - offset {
		length = memoryLen - offset
	}
	if length <= 0 {
		return
	}
	copy(memory[offset:], data)
}

// ReadHash reads a hash
func ReadHash(memory []byte, offset int) (*chainhash.Hash, error) {
	data := ReadBytes(memory, offset, chainhash.HashSize)
	if len(data) < chainhash.HashSize {
		return nil, fmt.Errorf("Memory out of bound")
	}

	return chainhash.NewHash(data)
}

// WriteHash writes a hash
func WriteHash(memory []byte, offset int, hash *chainhash.Hash) {
	WriteBytes(memory, offset, hash[:])
}

// WriteUint32 writes a uint32
func WriteUint32(memory []byte, offset int, value uint32) {
	endianess.PutUint32(memory[offset:], value)
}

// ReadUint32 reads a uint32
func ReadUint32(memory []byte, offset int) uint32 {
	return endianess.Uint32(memory[offset:])
}

// WriteUint64 writes a uint64
func WriteUint64(memory []byte, offset int, value uint64) {
	endianess.PutUint64(memory[offset:], value)
}

// ReadUint64 reads a uint64
func ReadUint64(memory []byte, offset int) uint64 {
	return endianess.Uint64(memory[offset:])
}

// WriteFloat32 writes a float32
func WriteFloat32(memory []byte, offset int, value float32) {
	WriteUint32(memory, offset, math.Float32bits(value))
}

// ReadFloat32 reads a float32
func ReadFloat32(memory []byte, offset int) float32 {
	return math.Float32frombits(ReadUint32(memory, offset))
}

// WriteFloat64 writes a float64
func WriteFloat64(memory []byte, offset int, value float64) {
	WriteUint64(memory, offset, math.Float64bits(value))
}

// ReadFloat64 reads a float64
func ReadFloat64(memory []byte, offset int) float64 {
	return math.Float64frombits(ReadUint64(memory, offset))
}

