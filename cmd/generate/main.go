// generate creates synthetic SSZ test data across structural axes.
//
// It produces a train/ and eval/ split so the optimization agent
// cannot overfit to a single payload shape.
//
// Usage: go run ./cmd/generate
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	mrand "math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ssz "github.com/pk910/dynamic-ssz"
	"github.com/prysmaticlabs/go-bitfield"

	. "github.com/pk910/autosszsearch"
)

// scenario defines one synthetic SSZ payload configuration.
type scenario struct {
	Name             string
	ValidatorCount   int
	AttestationCount int
	DepositCount     int
	TxCount          int
	TxMinSize        int
	TxMaxSize        int
	BlobCommitments  int
	SlashingCount    int
	VolExitCount     int
	Seed             int64
}

func main() {
	// Find res/ directory relative to source
	_, filename, _, _ := runtime.Caller(0)
	resDir := filepath.Join(filepath.Dir(filename), "..", "..", "res")

	trainDir := filepath.Join(resDir, "train")
	evalDir := filepath.Join(resDir, "eval")
	os.MkdirAll(trainDir, 0755)
	os.MkdirAll(evalDir, 0755)

	dynSsz := ssz.NewDynSsz(nil)

	// --- Training set ---
	// The agent benchmarks against these. Cover common structural patterns.
	trainScenarios := []scenario{
		{
			Name:             "block-full",
			ValidatorCount:   100000,
			AttestationCount: 128,
			DepositCount:     16,
			TxCount:          200,
			TxMinSize:        100,
			TxMaxSize:        800,
			BlobCommitments:  6,
			SlashingCount:    2,
			VolExitCount:     4,
			Seed:             1001,
		},
		{
			Name:             "block-empty",
			ValidatorCount:   100000,
			AttestationCount: 0,
			DepositCount:     0,
			TxCount:          0,
			TxMinSize:        0,
			TxMaxSize:        0,
			BlobCommitments:  0,
			SlashingCount:    0,
			VolExitCount:     0,
			Seed:             1002,
		},
		{
			Name:             "block-heavy-tx",
			ValidatorCount:   100000,
			AttestationCount: 10,
			DepositCount:     0,
			TxCount:          500,
			TxMinSize:        2000,
			TxMaxSize:        10000,
			BlobCommitments:  6,
			SlashingCount:    0,
			VolExitCount:     0,
			Seed:             1003,
		},
		{
			Name:             "state-small",
			ValidatorCount:   1000,
			AttestationCount: 0, // not used for state
			Seed:             2001,
		},
		{
			Name:             "state-medium",
			ValidatorCount:   50000,
			AttestationCount: 0,
			Seed:             2002,
		},
		{
			Name:             "state-large",
			ValidatorCount:   200000,
			AttestationCount: 0,
			Seed:             2003,
		},
	}

	// --- Evaluation set ---
	// The agent does NOT benchmark against these. Used to detect overfitting.
	evalScenarios := []scenario{
		{
			Name:             "block-mid-fill",
			ValidatorCount:   100000,
			AttestationCount: 64,
			DepositCount:     8,
			TxCount:          50,
			TxMinSize:        300,
			TxMaxSize:        1500,
			BlobCommitments:  3,
			SlashingCount:    1,
			VolExitCount:     2,
			Seed:             3001,
		},
		{
			Name:             "block-max-blobs",
			ValidatorCount:   100000,
			AttestationCount: 32,
			DepositCount:     0,
			TxCount:          100,
			TxMinSize:        500,
			TxMaxSize:        700,
			BlobCommitments:  32,
			SlashingCount:    0,
			VolExitCount:     0,
			Seed:             3002,
		},
		{
			Name:             "state-10k",
			ValidatorCount:   10000,
			AttestationCount: 0,
			Seed:             4001,
		},
		{
			Name:             "state-500k",
			ValidatorCount:   500000,
			AttestationCount: 0,
			Seed:             4002,
		},
	}

	fmt.Println("=== Generating training set ===")
	for _, s := range trainScenarios {
		if err := generateScenario(dynSsz, trainDir, s); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR [train/%s]: %v\n", s.Name, err)
			os.Exit(1)
		}
	}

	fmt.Println("\n=== Generating evaluation set ===")
	for _, s := range evalScenarios {
		if err := generateScenario(dynSsz, evalDir, s); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR [eval/%s]: %v\n", s.Name, err)
			os.Exit(1)
		}
	}

	fmt.Println("\nDone!")
}

func generateScenario(dynSsz *ssz.DynSsz, dir string, s scenario) error {
	rng := mrand.New(mrand.NewSource(s.Seed))

	isBlock := strings.HasPrefix(s.Name, "block")
	if isBlock {
		block := genBlock(rng, s)
		data, err := dynSsz.MarshalSSZ(block)
		if err != nil {
			return fmt.Errorf("marshal block: %w", err)
		}
		htr, err := dynSsz.HashTreeRoot(block.Message)
		if err != nil {
			return fmt.Errorf("block HTR: %w", err)
		}
		if err := writeFile(dir, s.Name+".ssz", data); err != nil {
			return err
		}
		if err := writeMeta(dir, s.Name+"-meta.json", htr); err != nil {
			return err
		}
		fmt.Printf("  %s/%s.ssz  (%d bytes, HTR: %s)\n", filepath.Base(dir), s.Name, len(data), hex.EncodeToString(htr[:]))
	}

	// Always generate a state if it's a state-* scenario
	if !isBlock {
		state := genState(rng, s)
		data, err := dynSsz.MarshalSSZ(state)
		if err != nil {
			return fmt.Errorf("marshal state: %w", err)
		}
		htr, err := dynSsz.HashTreeRoot(state)
		if err != nil {
			return fmt.Errorf("state HTR: %w", err)
		}
		if err := writeFile(dir, s.Name+".ssz", data); err != nil {
			return err
		}
		if err := writeMeta(dir, s.Name+"-meta.json", htr); err != nil {
			return err
		}
		fmt.Printf("  %s/%s.ssz  (%d bytes, HTR: %s)\n", filepath.Base(dir), s.Name, len(data), hex.EncodeToString(htr[:]))
	}

	return nil
}

func writeFile(dir, name string, data []byte) error {
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}

func writeMeta(dir, name string, htr [32]byte) error {
	meta := Metadata{HTR: hex.EncodeToString(htr[:])}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}

// === Block generation ===

func genBlock(rng *mrand.Rand, s scenario) *SignedBeaconBlock {
	return &SignedBeaconBlock{
		Message: &BeaconBlock{
			Slot:          uint64(rng.Int63n(1000000)),
			ProposerIndex: uint64(rng.Int63n(int64(s.ValidatorCount))),
			ParentRoot:    randRoot(rng),
			StateRoot:     randRoot(rng),
			Body:          genBlockBody(rng, s),
		},
		Signature: randBLSSig(rng),
	}
}

func genBlockBody(rng *mrand.Rand, s scenario) *BeaconBlockBody {
	return &BeaconBlockBody{
		RANDAOReveal:          randBLSSig(rng),
		ETH1Data:              genETH1Data(rng),
		Graffiti:              randHash32(rng),
		ProposerSlashings:     genProposerSlashings(rng, s.SlashingCount, s.ValidatorCount),
		AttesterSlashings:     genAttesterSlashings(rng, s.SlashingCount, s.ValidatorCount),
		Attestations:          genAttestations(rng, s.AttestationCount, s.ValidatorCount),
		Deposits:              genDeposits(rng, s.DepositCount),
		VoluntaryExits:        genVolExits(rng, s.VolExitCount, s.ValidatorCount),
		SyncAggregate:         genSyncAggregate(rng),
		ExecutionPayload:      genExecPayload(rng, s),
		BLSToExecutionChanges: genBLSChanges(rng, min(s.VolExitCount, 16), s.ValidatorCount),
		BlobKZGCommitments:    genKZGCommitments(rng, s.BlobCommitments),
	}
}

// === State generation ===

func genState(rng *mrand.Rand, s scenario) *BeaconState {
	n := s.ValidatorCount
	validators := make([]*Validator, n)
	balances := make([]Gwei, n)
	prevPart := make([]ParticipationFlags, n)
	currPart := make([]ParticipationFlags, n)
	inactivity := make([]uint64, n)

	for i := 0; i < n; i++ {
		validators[i] = &Validator{
			Pubkey:                     randBLSPubKey(rng),
			WithdrawalCredentials:      randHash32(rng),
			EffectiveBalance:           32000000000,
			Slashed:                    false,
			ActivationEligibilityEpoch: 0,
			ActivationEpoch:            0,
			ExitEpoch:                  ^uint64(0),
			WithdrawableEpoch:          ^uint64(0),
		}
		balances[i] = 32000000000 + Gwei(rng.Int63n(1000000000))
		prevPart[i] = ParticipationFlags(rng.Intn(8))
		currPart[i] = ParticipationFlags(rng.Intn(8))
		inactivity[i] = uint64(rng.Int63n(100))
	}

	blockRoots := make([]Root, 8192)
	stateRoots := make([]Root, 8192)
	for i := range blockRoots {
		blockRoots[i] = randRoot(rng)
		stateRoots[i] = randRoot(rng)
	}

	randaoMixes := make([]Root, 65536)
	for i := range randaoMixes {
		randaoMixes[i] = randRoot(rng)
	}

	slashings := make([]Gwei, 8192)
	for i := range slashings {
		slashings[i] = Gwei(rng.Int63n(32000000000))
	}

	eth1VoteCount := 32 * 64 // SLOTS_PER_EPOCH * EPOCHS_PER_ETH1_VOTING
	eth1Votes := make([]*ETH1Data, eth1VoteCount)
	for i := range eth1Votes {
		eth1Votes[i] = genETH1Data(rng)
	}

	slot := uint64(rng.Int63n(1000000))
	epoch := slot / 32

	txRoot := sha256.Sum256(randBytes(rng, 32))
	wdRoot := sha256.Sum256(randBytes(rng, 32))

	return &BeaconState{
		GenesisTime:           1606824023,
		GenesisValidatorsRoot: randRoot(rng),
		Slot:                  slot,
		Fork: &Fork{
			PreviousVersion: [4]byte{0x04, 0x00, 0x00, 0x00},
			CurrentVersion:  [4]byte{0x04, 0x00, 0x00, 0x00},
			Epoch:           epoch,
		},
		LatestBlockHeader: &BeaconBlockHeader{
			Slot:          slot - 1,
			ProposerIndex: uint64(rng.Int63n(int64(n))),
			ParentRoot:    randRoot(rng),
			StateRoot:     randRoot(rng),
			BodyRoot:      randRoot(rng),
		},
		BlockRoots:      blockRoots,
		StateRoots:      stateRoots,
		HistoricalRoots: []Root{},
		ETH1Data:        genETH1Data(rng),
		ETH1DataVotes:   eth1Votes,
		ETH1DepositIndex:           uint64(n),
		Validators:                 validators,
		Balances:                   balances,
		RANDAOMixes:                randaoMixes,
		Slashings:                  slashings,
		PreviousEpochParticipation: prevPart,
		CurrentEpochParticipation:  currPart,
		JustificationBits:          bitfield.NewBitvector4(),
		PreviousJustifiedCheckpoint: &Checkpoint{
			Epoch: epoch - 2,
			Root:  randRoot(rng),
		},
		CurrentJustifiedCheckpoint: &Checkpoint{
			Epoch: epoch - 1,
			Root:  randRoot(rng),
		},
		FinalizedCheckpoint: &Checkpoint{
			Epoch: epoch - 2,
			Root:  randRoot(rng),
		},
		InactivityScores:     inactivity,
		CurrentSyncCommittee: genSyncCommittee(rng),
		NextSyncCommittee:    genSyncCommittee(rng),
		LatestExecutionPayloadHeader: &ExecutionPayloadHeader{
			ParentHash:       randHash32(rng),
			FeeRecipient:     randExecAddr(rng),
			StateRoot:        randHash32(rng),
			ReceiptsRoot:     randHash32(rng),
			LogsBloom:        randLogsBloom(rng),
			PrevRandao:       randHash32(rng),
			BlockNumber:      uint64(rng.Int63n(10000000)),
			GasLimit:         30000000,
			GasUsed:          uint64(15000000 + rng.Int63n(10000000)),
			Timestamp:        uint64(1700000000 + rng.Int63n(10000000)),
			ExtraData:        randBytes(rng, 32),
			BaseFeePerGas:    randUint256(rng),
			BlockHash:        randHash32(rng),
			TransactionsRoot: txRoot,
			WithdrawalsRoot:  wdRoot,
			BlobGasUsed:      uint64(rng.Int63n(1000000)),
			ExcessBlobGas:    uint64(rng.Int63n(1000000)),
		},
		NextWithdrawalIndex:          uint64(rng.Int63n(1000000)),
		NextWithdrawalValidatorIndex: uint64(rng.Int63n(int64(n))),
		HistoricalSummaries:          []*HistoricalSummary{},
	}
}

// === Sub-type generators ===

func genETH1Data(rng *mrand.Rand) *ETH1Data {
	return &ETH1Data{
		DepositRoot:  randRoot(rng),
		DepositCount: uint64(rng.Int63n(1000000)),
		BlockHash:    randHash32(rng),
	}
}

func genProposerSlashings(rng *mrand.Rand, count, valCount int) []*ProposerSlashing {
	out := make([]*ProposerSlashing, count)
	for i := range out {
		vi := uint64(rng.Int63n(int64(valCount)))
		out[i] = &ProposerSlashing{
			SignedHeader1: &SignedBeaconBlockHeader{
				Message: &BeaconBlockHeader{
					Slot: uint64(rng.Int63n(1000)), ProposerIndex: vi,
					ParentRoot: randRoot(rng), StateRoot: randRoot(rng), BodyRoot: randRoot(rng),
				},
				Signature: randBLSSig(rng),
			},
			SignedHeader2: &SignedBeaconBlockHeader{
				Message: &BeaconBlockHeader{
					Slot: uint64(rng.Int63n(1000)), ProposerIndex: vi,
					ParentRoot: randRoot(rng), StateRoot: randRoot(rng), BodyRoot: randRoot(rng),
				},
				Signature: randBLSSig(rng),
			},
		}
	}
	return out
}

func genAttesterSlashings(rng *mrand.Rand, count, valCount int) []*AttesterSlashing {
	out := make([]*AttesterSlashing, count)
	for i := range out {
		n := 10 + rng.Intn(50)
		indices := make([]uint64, n)
		for j := range indices {
			indices[j] = uint64(rng.Int63n(int64(valCount)))
		}
		out[i] = &AttesterSlashing{
			Attestation1: &IndexedAttestation{
				AttestingIndices: indices, Data: genAttData(rng), Signature: randBLSSig(rng),
			},
			Attestation2: &IndexedAttestation{
				AttestingIndices: indices, Data: genAttData(rng), Signature: randBLSSig(rng),
			},
		}
	}
	return out
}

func genAttData(rng *mrand.Rand) *AttestationData {
	slot := uint64(rng.Int63n(1000000))
	epoch := slot / 32
	return &AttestationData{
		Slot: slot, Index: uint64(rng.Int63n(64)), BeaconBlockRoot: randRoot(rng),
		Source: &Checkpoint{Epoch: epoch - 1, Root: randRoot(rng)},
		Target: &Checkpoint{Epoch: epoch, Root: randRoot(rng)},
	}
}

func genAttestations(rng *mrand.Rand, count, valCount int) []*Attestation {
	out := make([]*Attestation, count)
	for i := range out {
		numBits := 64 + rng.Intn(200)
		bits := bitfield.NewBitlist(uint64(numBits))
		for j := 0; j < numBits*2/3; j++ {
			bits.SetBitAt(uint64(j), true)
		}
		out[i] = &Attestation{
			AggregationBits: bits, Data: genAttData(rng), Signature: randBLSSig(rng),
		}
	}
	return out
}

func genDeposits(rng *mrand.Rand, count int) []*Deposit {
	out := make([]*Deposit, count)
	for i := range out {
		proof := make([][]byte, 33)
		for j := range proof {
			proof[j] = randBytes(rng, 32)
		}
		out[i] = &Deposit{
			Proof: proof,
			Data: &DepositData{
				Pubkey: randBLSPubKey(rng), WithdrawalCredentials: randHash32(rng),
				Amount: 32000000000, Signature: randBLSSig(rng),
			},
		}
	}
	return out
}

func genVolExits(rng *mrand.Rand, count, valCount int) []*SignedVoluntaryExit {
	out := make([]*SignedVoluntaryExit, count)
	for i := range out {
		out[i] = &SignedVoluntaryExit{
			Message: &VoluntaryExit{
				Epoch: uint64(rng.Int63n(1000)), ValidatorIndex: uint64(rng.Int63n(int64(valCount))),
			},
			Signature: randBLSSig(rng),
		}
	}
	return out
}

func genSyncAggregate(rng *mrand.Rand) *SyncAggregate {
	var bits bitfield.Bitvector512
	for i := 0; i < 512*2/3; i++ {
		bits.SetBitAt(uint64(i), true)
	}
	return &SyncAggregate{SyncCommitteeBits: bits, SyncCommitteeSignature: randBLSSig(rng)}
}

func genSyncCommittee(rng *mrand.Rand) *SyncCommittee {
	pks := make([]BLSPubKey, 512)
	for i := range pks {
		pks[i] = randBLSPubKey(rng)
	}
	return &SyncCommittee{Pubkeys: pks, AggregatePubkey: randBLSPubKey(rng)}
}

func genExecPayload(rng *mrand.Rand, s scenario) *ExecutionPayload {
	txns := make([][]byte, s.TxCount)
	for i := range txns {
		sz := s.TxMinSize
		if s.TxMaxSize > s.TxMinSize {
			sz += rng.Intn(s.TxMaxSize - s.TxMinSize)
		}
		txns[i] = randBytes(rng, sz)
	}
	wds := make([]*Withdrawal, min(16, max(1, s.TxCount/50)))
	if s.TxCount == 0 {
		wds = nil
	}
	for i := range wds {
		wds[i] = &Withdrawal{
			Index: uint64(i), ValidatorIndex: uint64(rng.Int63n(int64(s.ValidatorCount))),
			Address: randExecAddr(rng), Amount: Gwei(rng.Int63n(32000000000)),
		}
	}
	return &ExecutionPayload{
		ParentHash: randHash32(rng), FeeRecipient: randExecAddr(rng),
		StateRoot: randHash32(rng), ReceiptsRoot: randHash32(rng),
		LogsBloom: randLogsBloom(rng), PrevRandao: randHash32(rng),
		BlockNumber: uint64(rng.Int63n(10000000)), GasLimit: 30000000,
		GasUsed: uint64(15000000 + rng.Int63n(10000000)),
		Timestamp: uint64(1700000000 + rng.Int63n(10000000)),
		ExtraData: randBytes(rng, 32), BaseFeePerGas: randUint256(rng),
		BlockHash: randHash32(rng), Transactions: txns, Withdrawals: wds,
		BlobGasUsed: uint64(rng.Int63n(1000000)), ExcessBlobGas: uint64(rng.Int63n(1000000)),
	}
}

func genBLSChanges(rng *mrand.Rand, count, valCount int) []*SignedBLSToExecutionChange {
	out := make([]*SignedBLSToExecutionChange, count)
	for i := range out {
		out[i] = &SignedBLSToExecutionChange{
			Message: &BLSToExecutionChange{
				ValidatorIndex: uint64(rng.Int63n(int64(valCount))),
				FromBLSPubkey:  randBLSPubKey(rng), ToExecutionAddress: randExecAddr(rng),
			},
			Signature: randBLSSig(rng),
		}
	}
	return out
}

func genKZGCommitments(rng *mrand.Rand, count int) []KZGCommitment {
	out := make([]KZGCommitment, count)
	for i := range out {
		rng.Read(out[i][:])
	}
	return out
}

// === Random primitives (deterministic via seeded rng) ===

func randBytes(rng *mrand.Rand, n int) []byte {
	b := make([]byte, n)
	rng.Read(b)
	return b
}

func randRoot(rng *mrand.Rand) Root {
	var r Root
	rng.Read(r[:])
	return r
}

func randHash32(rng *mrand.Rand) Hash32 {
	var h Hash32
	rng.Read(h[:])
	return h
}

func randBLSPubKey(rng *mrand.Rand) BLSPubKey {
	var k BLSPubKey
	rng.Read(k[:])
	return k
}

func randBLSSig(rng *mrand.Rand) BLSSignature {
	var s BLSSignature
	rng.Read(s[:])
	return s
}

func randExecAddr(rng *mrand.Rand) ExecutionAddress {
	var a ExecutionAddress
	rng.Read(a[:])
	return a
}

func randLogsBloom(rng *mrand.Rand) LogsBloom {
	var l LogsBloom
	rng.Read(l[:])
	return l
}

func randUint256(rng *mrand.Rand) Uint256 {
	var u Uint256
	rng.Read(u[:])
	return u
}

// Ensure crypto/rand and big are used (needed for compilation)
var _ = rand.Reader
var _ = new(big.Int)
