package autosszsearch

// api.go — Thin wrappers calling into the dynamic-ssz library.
//
// The agent MAY modify this file to wire up new library APIs (e.g. view types,
// lazy accessors) to the benchmarks. However, all decoding/encoding logic MUST
// live in the dynamic-ssz/ library, not here. This file should contain only
// function calls into the library, not reimplementations.

import (
	ssz "github.com/pk910/dynamic-ssz"
)

var dynSsz = ssz.NewDynSsz(nil)

// --- Unmarshal ---

func UnmarshalBlock(data []byte) (*SignedBeaconBlock, error) {
	block := new(SignedBeaconBlock)
	if err := dynSsz.UnmarshalSSZ(block, data); err != nil {
		return nil, err
	}
	return block, nil
}

func UnmarshalState(data []byte) (*BeaconState, error) {
	state := new(BeaconState)
	if err := dynSsz.UnmarshalSSZ(state, data); err != nil {
		return nil, err
	}
	return state, nil
}

// --- Marshal ---

func MarshalBlock(block *SignedBeaconBlock) ([]byte, error) {
	return dynSsz.MarshalSSZ(block)
}

func MarshalState(state *BeaconState) ([]byte, error) {
	return dynSsz.MarshalSSZ(state)
}

// --- HashTreeRoot ---

func HashTreeRootBlock(block *BeaconBlock) ([32]byte, error) {
	return dynSsz.HashTreeRoot(block)
}

func HashTreeRootState(state *BeaconState) ([32]byte, error) {
	return dynSsz.HashTreeRoot(state)
}

// --- Selective field access ---
// Baseline: full unmarshal then read field.
// Agent should replace these with library-level zero-copy APIs.

func ReadBlockSlot(data []byte) (uint64, error) {
	block, err := UnmarshalBlock(data)
	if err != nil {
		return 0, err
	}
	return block.Message.Slot, nil
}

func ReadStateSlot(data []byte) (uint64, error) {
	state, err := UnmarshalState(data)
	if err != nil {
		return 0, err
	}
	return state.Slot, nil
}

func ReadStateValidatorCount(data []byte) (int, error) {
	state, err := UnmarshalState(data)
	if err != nil {
		return 0, err
	}
	return len(state.Validators), nil
}

func ReadStateBalance(data []byte, validatorIndex int) (uint64, error) {
	state, err := UnmarshalState(data)
	if err != nil {
		return 0, err
	}
	if validatorIndex >= len(state.Balances) {
		return 0, nil
	}
	return state.Balances[validatorIndex], nil
}
