package autosszsearch

import (
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
