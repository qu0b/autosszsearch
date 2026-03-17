package autosszsearch

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Metadata holds the pre-computed hash tree root for validation
type Metadata struct {
	HTR string `json:"htr"`
}

// TestPayload holds one SSZ test file with its expected HTR.
type TestPayload struct {
	Name string
	Data []byte
	HTR  [32]byte
}

var (
	// Original mainnet data (backward compat)
	blockMainnetData []byte
	stateMainnetData []byte
	blockMainnetHTR  [32]byte
	stateMainnetHTR  [32]byte

	// Training set: agent benchmarks against these
	TrainBlocks []TestPayload
	TrainStates []TestPayload

	// Evaluation set: used to detect overfitting (not benchmarked by agent)
	EvalBlocks []TestPayload
	EvalStates []TestPayload

	resDir string
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	resDir = filepath.Join(filepath.Dir(filename), "res")

	// Load original mainnet data
	blockMainnetData = mustReadFile(filepath.Join(resDir, "block-mainnet.ssz"))
	stateMainnetData = mustReadFile(filepath.Join(resDir, "state-mainnet.ssz"))
	blockMainnetHTR = loadHTR(filepath.Join(resDir, "block-mainnet-meta.json"))
	stateMainnetHTR = loadHTR(filepath.Join(resDir, "state-mainnet-meta.json"))

	// Load train/eval sets (non-fatal if missing — generator may not have run)
	TrainBlocks, TrainStates = loadPayloadSet(filepath.Join(resDir, "train"))
	EvalBlocks, EvalStates = loadPayloadSet(filepath.Join(resDir, "eval"))
}

func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("failed to load " + path + ": " + err.Error())
	}
	return data
}

func loadHTR(path string) [32]byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("failed to load " + path + ": " + err.Error())
	}
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		panic("failed to parse " + path + ": " + err.Error())
	}
	htrBytes, err := hex.DecodeString(meta.HTR)
	if err != nil {
		panic("failed to decode HTR from " + path + ": " + err.Error())
	}
	var htr [32]byte
	copy(htr[:], htrBytes)
	return htr
}

func loadPayloadSet(dir string) (blocks, states []TestPayload) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil // dir doesn't exist, that's fine
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".ssz") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".ssz")
		metaPath := filepath.Join(dir, name+"-meta.json")

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping %s: %v\n", e.Name(), err)
			continue
		}
		htr := loadHTR(metaPath)

		p := TestPayload{Name: name, Data: data, HTR: htr}
		if strings.HasPrefix(name, "block") {
			blocks = append(blocks, p)
		} else {
			states = append(states, p)
		}
	}
	return blocks, states
}

// --- Accessors (backward compat) ---

func GetBlockMainnetData() []byte    { return blockMainnetData }
func GetStateMainnetData() []byte    { return stateMainnetData }
func GetBlockMainnetHTR() [32]byte   { return blockMainnetHTR }
func GetStateMainnetHTR() [32]byte   { return stateMainnetHTR }
