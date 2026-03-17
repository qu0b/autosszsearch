package autosszsearch

import (
	"bytes"
	"testing"
)

// ========================= BLOCK MAINNET BENCHMARKS =========================

func BenchmarkBlockMainnet_Unmarshal(b *testing.B) {
	data := GetBlockMainnetData()
	var block *SignedBeaconBlock
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		block, err = UnmarshalBlock(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	htr, err := HashTreeRootBlock(block.Message)
	if err != nil {
		b.Fatal(err)
	}
	expected := GetBlockMainnetHTR()
	if htr != expected {
		b.Fatalf("HTR mismatch: got %x, want %x", htr, expected)
	}
}

func BenchmarkBlockMainnet_Marshal(b *testing.B) {
	data := GetBlockMainnetData()
	block, err := UnmarshalBlock(data)
	if err != nil {
		b.Fatal(err)
	}
	var out []byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err = MarshalBlock(block)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if !bytes.Equal(out, data) {
		b.Fatal("marshal round-trip mismatch")
	}
}

func BenchmarkBlockMainnet_HashTreeRoot(b *testing.B) {
	data := GetBlockMainnetData()
	block, err := UnmarshalBlock(data)
	if err != nil {
		b.Fatal(err)
	}
	var htr [32]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		htr, err = HashTreeRootBlock(block.Message)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	expected := GetBlockMainnetHTR()
	if htr != expected {
		b.Fatalf("HTR mismatch: got %x, want %x", htr, expected)
	}
}

// ========================= STATE MAINNET BENCHMARKS =========================

func BenchmarkStateMainnet_Unmarshal(b *testing.B) {
	data := GetStateMainnetData()
	var state *BeaconState
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		state, err = UnmarshalState(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	htr, err := HashTreeRootState(state)
	if err != nil {
		b.Fatal(err)
	}
	expected := GetStateMainnetHTR()
	if htr != expected {
		b.Fatalf("HTR mismatch: got %x, want %x", htr, expected)
	}
}

func BenchmarkStateMainnet_Marshal(b *testing.B) {
	data := GetStateMainnetData()
	state, err := UnmarshalState(data)
	if err != nil {
		b.Fatal(err)
	}
	var out []byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err = MarshalState(state)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if !bytes.Equal(out, data) {
		b.Fatal("marshal round-trip mismatch")
	}
}

func BenchmarkStateMainnet_HashTreeRoot(b *testing.B) {
	data := GetStateMainnetData()
	state, err := UnmarshalState(data)
	if err != nil {
		b.Fatal(err)
	}
	var htr [32]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		htr, err = HashTreeRootState(state)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	expected := GetStateMainnetHTR()
	if htr != expected {
		b.Fatalf("HTR mismatch: got %x, want %x", htr, expected)
	}
}

// ========================= SELECTIVE FIELD ACCESS ===========================

func BenchmarkBlockMainnet_ReadSlot(b *testing.B) {
	data := GetBlockMainnetData()
	block, err := UnmarshalBlock(data)
	if err != nil {
		b.Fatal(err)
	}
	expectedSlot := block.Message.Slot

	var slot uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slot, err = ReadBlockSlot(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if slot != expectedSlot {
		b.Fatalf("slot mismatch: got %d, want %d", slot, expectedSlot)
	}
}

func BenchmarkStateMainnet_ReadSlot(b *testing.B) {
	data := GetStateMainnetData()
	state, err := UnmarshalState(data)
	if err != nil {
		b.Fatal(err)
	}
	expectedSlot := state.Slot

	var slot uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slot, err = ReadStateSlot(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if slot != expectedSlot {
		b.Fatalf("slot mismatch: got %d, want %d", slot, expectedSlot)
	}
}

func BenchmarkStateMainnet_ReadValidatorCount(b *testing.B) {
	data := GetStateMainnetData()
	state, err := UnmarshalState(data)
	if err != nil {
		b.Fatal(err)
	}
	expectedCount := len(state.Validators)

	var count int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count, err = ReadStateValidatorCount(data)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if count != expectedCount {
		b.Fatalf("validator count mismatch: got %d, want %d", count, expectedCount)
	}
}

func BenchmarkStateMainnet_ReadBalance(b *testing.B) {
	data := GetStateMainnetData()
	state, err := UnmarshalState(data)
	if err != nil {
		b.Fatal(err)
	}
	targetIdx := len(state.Balances) / 2
	expectedBalance := state.Balances[targetIdx]

	var balance uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		balance, err = ReadStateBalance(data, targetIdx)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	if balance != expectedBalance {
		b.Fatalf("balance mismatch: got %d, want %d", balance, expectedBalance)
	}
}
