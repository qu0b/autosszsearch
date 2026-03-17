package autosszsearch

import (
	"bytes"
	"fmt"
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

// ========================= TRAINING SET BENCHMARKS ==========================
// These run unmarshal + HTR round-trip across all training payloads.
// Ensures optimizations generalize beyond the single mainnet payload.

func BenchmarkTrain_BlockUnmarshal(b *testing.B) {
	if len(TrainBlocks) == 0 {
		b.Skip("no training blocks (run: go run ./cmd/generate)")
	}
	for _, p := range TrainBlocks {
		b.Run(p.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := UnmarshalBlock(p.Data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTrain_StateUnmarshal(b *testing.B) {
	if len(TrainStates) == 0 {
		b.Skip("no training states (run: go run ./cmd/generate)")
	}
	for _, p := range TrainStates {
		b.Run(p.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := UnmarshalState(p.Data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTrain_BlockHTR(b *testing.B) {
	if len(TrainBlocks) == 0 {
		b.Skip("no training blocks (run: go run ./cmd/generate)")
	}
	for _, p := range TrainBlocks {
		b.Run(p.Name, func(b *testing.B) {
			block, err := UnmarshalBlock(p.Data)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				htr, err := HashTreeRootBlock(block.Message)
				if err != nil {
					b.Fatal(err)
				}
				if htr != p.HTR {
					b.Fatalf("HTR mismatch for %s", p.Name)
				}
			}
		})
	}
}

func BenchmarkTrain_StateHTR(b *testing.B) {
	if len(TrainStates) == 0 {
		b.Skip("no training states (run: go run ./cmd/generate)")
	}
	for _, p := range TrainStates {
		b.Run(p.Name, func(b *testing.B) {
			state, err := UnmarshalState(p.Data)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				htr, err := HashTreeRootState(state)
				if err != nil {
					b.Fatal(err)
				}
				if htr != p.HTR {
					b.Fatalf("HTR mismatch for %s", p.Name)
				}
			}
		})
	}
}

// ========================= EVALUATION SET (test, not benchmark) =============
// TestEval verifies correctness on the held-out evaluation set.
// Run after experiments: go test -run=TestEval -v

func TestEvalBlocks(t *testing.T) {
	if len(EvalBlocks) == 0 {
		t.Skip("no eval blocks (run: go run ./cmd/generate)")
	}
	for _, p := range EvalBlocks {
		t.Run(p.Name, func(t *testing.T) {
			block, err := UnmarshalBlock(p.Data)
			if err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			htr, err := HashTreeRootBlock(block.Message)
			if err != nil {
				t.Fatalf("HTR failed: %v", err)
			}
			if htr != p.HTR {
				t.Fatalf("HTR mismatch: got %x, want %x", htr, p.HTR)
			}
			data, err := MarshalBlock(block)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			if !bytes.Equal(data, p.Data) {
				t.Fatalf("round-trip mismatch (len got=%d want=%d)", len(data), len(p.Data))
			}
			fmt.Printf("  %s: OK (%d bytes)\n", p.Name, len(p.Data))
		})
	}
}

func TestEvalStates(t *testing.T) {
	if len(EvalStates) == 0 {
		t.Skip("no eval states (run: go run ./cmd/generate)")
	}
	for _, p := range EvalStates {
		t.Run(p.Name, func(t *testing.T) {
			state, err := UnmarshalState(p.Data)
			if err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			htr, err := HashTreeRootState(state)
			if err != nil {
				t.Fatalf("HTR failed: %v", err)
			}
			if htr != p.HTR {
				t.Fatalf("HTR mismatch: got %x, want %x", htr, p.HTR)
			}
			data, err := MarshalState(state)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			if !bytes.Equal(data, p.Data) {
				t.Fatalf("round-trip mismatch (len got=%d want=%d)", len(data), len(p.Data))
			}
			fmt.Printf("  %s: OK (%d bytes)\n", p.Name, len(p.Data))
		})
	}
}
