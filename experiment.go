package autosszsearch

// experiment.go — THE ONLY FILE THE AGENT MODIFIES
//
// This file contains SSZ optimization implementations.
// The agent iterates on this file to find faster approaches
// to SSZ deserialization, with a focus on zero-copy techniques
// inspired by Rust's rkyv library.
//
// Current approach: baseline using dynamic-ssz standard unmarshal.

import (
	ssz "github.com/pk910/dynamic-ssz"
)

var dynSsz = ssz.NewDynSsz(nil)

// UnmarshalBlock deserializes a SignedBeaconBlock from SSZ bytes.
// This is the primary function the agent optimizes.
func UnmarshalBlock(data []byte) (*SignedBeaconBlock, error) {
	block := new(SignedBeaconBlock)
	if err := dynSsz.UnmarshalSSZ(block, data); err != nil {
		return nil, err
	}
	return block, nil
}

// UnmarshalState deserializes a BeaconState from SSZ bytes.
// This is the secondary function the agent optimizes.
func UnmarshalState(data []byte) (*BeaconState, error) {
	state := new(BeaconState)
	if err := dynSsz.UnmarshalSSZ(state, data); err != nil {
		return nil, err
	}
	return state, nil
}

// HashTreeRootBlock computes the hash tree root of a BeaconBlock.
func HashTreeRootBlock(block *BeaconBlock) ([32]byte, error) {
	return dynSsz.HashTreeRoot(block)
}

// HashTreeRootState computes the hash tree root of a BeaconState.
func HashTreeRootState(state *BeaconState) ([32]byte, error) {
	return dynSsz.HashTreeRoot(state)
}
