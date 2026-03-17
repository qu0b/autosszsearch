package autosszsearch

// api.go — Thin wrappers calling into the dynamic-ssz library.
//
// The agent MAY modify this file to wire up new library APIs (e.g. view types,
// lazy accessors) to the benchmarks. However, all decoding/encoding logic MUST
// live in the dynamic-ssz/ library, not here. This file should contain only
// function calls into the library, not reimplementations.

import (
	"reflect"

	ssz "github.com/pk910/dynamic-ssz"
)

var dynSsz = ssz.NewDynSsz(nil)

// Type handles for zero-copy field access
var (
	signedBeaconBlockType = reflect.TypeOf((*SignedBeaconBlock)(nil))
	beaconStateType       = reflect.TypeOf((*BeaconState)(nil))
)

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
// Zero-copy: navigate SSZ offsets to read fields without full unmarshal.

func ReadBlockSlot(data []byte) (uint64, error) {
	// SignedBeaconBlock.Message(0).Slot(0)
	return dynSsz.ReadUint64(signedBeaconBlockType, data, 0, 0)
}

func ReadStateSlot(data []byte) (uint64, error) {
	// BeaconState.Slot(2)
	return dynSsz.ReadUint64(beaconStateType, data, 2)
}

func ReadStateValidatorCount(data []byte) (int, error) {
	// BeaconState.Validators(11)
	return dynSsz.ReadListLength(beaconStateType, data, 11)
}

func ReadStateBalance(data []byte, validatorIndex int) (uint64, error) {
	// BeaconState.Balances(12)
	return dynSsz.ReadUint64FromList(beaconStateType, data, validatorIndex, 12)
}
