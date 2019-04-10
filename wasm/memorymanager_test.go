package wasm_test

import (
	"testing"

	"github.com/phoreproject/synapse/wasm"
)

func TestMemoryManager(t *testing.T) {
	top := 8
	buffer := make([]byte, 1024)
	mm := wasm.NewMemoryManagerForBuffer(buffer, top)

	addr1 := mm.Allocate(1)

	if addr1 != top {
		t.Fatalf("MM allocates at wrong address %d, expected %d", addr1, top)
	}

	addr2 := mm.Allocate(4)

	if addr2 != top+1 {
		t.Fatalf("MM allocates at wrong address %d, expected %d", addr2, top+1)
	}

	mm.Free(addr2)
	addr2 = mm.Allocate(4)

	if addr2 != top+1 {
		t.Fatalf("MM allocates at wrong address %d, expected %d", addr2, top+1)
	}

	addr3 := mm.Allocate(8)

	if addr3 != top+5 {
		t.Fatalf("MM allocates at wrong address %d, expected %d", addr3, top+5)
	}

	mm.Free(addr2)
	mm.Free(addr1)
	mm.Free(addr3)

	addr1 = mm.Allocate(1)

	if addr1 != top {
		t.Fatalf("MM allocates at wrong address %d, expected %d", addr1, top)
	}
}
