package autosszsearch

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Metadata holds the pre-computed hash tree root for validation
type Metadata struct {
	HTR string `json:"htr"`
}

var (
	blockMainnetData []byte
	stateMainnetData []byte

	blockMainnetHTR [32]byte
	stateMainnetHTR [32]byte

	resDir string
)

func init() {
	// Find the res directory relative to this source file
	_, filename, _, _ := runtime.Caller(0)
	resDir = filepath.Join(filepath.Dir(filename), "res")

	var err error

	// Load test data
	blockMainnetData, err = os.ReadFile(filepath.Join(resDir, "block-mainnet.ssz"))
	if err != nil {
		panic("failed to load block-mainnet.ssz: " + err.Error())
	}
	stateMainnetData, err = os.ReadFile(filepath.Join(resDir, "state-mainnet.ssz"))
	if err != nil {
		panic("failed to load state-mainnet.ssz: " + err.Error())
	}

	// Load metadata (pre-computed HTRs for correctness validation)
	blockMainnetHTR = loadHTR(filepath.Join(resDir, "block-mainnet-meta.json"))
	stateMainnetHTR = loadHTR(filepath.Join(resDir, "state-mainnet-meta.json"))
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

// GetBlockMainnetData returns the raw SSZ-encoded mainnet block bytes.
func GetBlockMainnetData() []byte {
	return blockMainnetData
}

// GetStateMainnetData returns the raw SSZ-encoded mainnet state bytes.
func GetStateMainnetData() []byte {
	return stateMainnetData
}

// GetBlockMainnetHTR returns the pre-computed hash tree root for the mainnet block.
func GetBlockMainnetHTR() [32]byte {
	return blockMainnetHTR
}

// GetStateMainnetHTR returns the pre-computed hash tree root for the mainnet state.
func GetStateMainnetHTR() [32]byte {
	return stateMainnetHTR
}
