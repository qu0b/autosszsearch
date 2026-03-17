# autosszsearch

Autonomous research to improve [dynamic-ssz](https://github.com/pk910/dynamic-ssz) library performance.

Inspired by [rkyv](https://github.com/rkyv/rkyv)'s zero-copy deserialization, this project uses the [autoresearch](https://github.com/qu0b/autoresearch) pattern to autonomously discover SSZ optimization techniques that can be upstreamed into the library.

## How it works

An AI agent runs in a loop:
1. Modifies the **dynamic-ssz library** source code (included as a local fork in `dynamic-ssz/`)
2. Runs library tests to verify correctness
3. Runs benchmarks against real Ethereum mainnet data
4. Keeps improvements, discards regressions
5. Logs all results to `results.tsv`
6. Repeats indefinitely

The key difference from a standalone experiment: **all optimizations live in the library**, not in bespoke one-off decoders. Every improvement is generalizable and upstreamable.

## Structure

```
autosszsearch/
├── program.md          # Agent instructions (the experiment loop)
├── api.go              # Thin wrappers calling into the library (agent may modify)
├── bench_test.go       # Fixed benchmark harness (DO NOT MODIFY)
├── prepare.go          # Fixed test data loading (DO NOT MODIFY)
├── types.go            # Ethereum Deneb types (DO NOT MODIFY)
├── dynamic-ssz/        # Fork of the library — agent modifies files here
│   ├── dynssz.go       # Main API
│   ├── reflection/     # Runtime marshal/unmarshal/hash
│   ├── sszutils/       # Encoder/decoder primitives
│   ├── codegen/        # Code generation
│   ├── hasher/         # Incremental hasher (PR #133)
│   └── ...
├── res/                # Real Ethereum mainnet test data
│   ├── block-mainnet.ssz       (127KB)
│   └── state-mainnet.ssz       (17MB)
└── results.tsv         # Experiment log (untracked)
```

## Baseline

The library fork is based on [PR #133](https://github.com/pk910/dynamic-ssz/pull/133) (incremental hasher).

## Benchmarks

| Benchmark | What it measures |
|-----------|-----------------|
| `BlockMainnet_Unmarshal` | Deserialize 127KB beacon block |
| `StateMainnet_Unmarshal` | Deserialize 17MB beacon state |
| `BlockMainnet_Marshal` | Serialize beacon block back to bytes |
| `StateMainnet_Marshal` | Serialize beacon state back to bytes |
| `BlockMainnet_HashTreeRoot` | Merkle root of beacon block |
| `StateMainnet_HashTreeRoot` | Merkle root of beacon state |
| `BlockMainnet_ReadSlot` | Read single field from raw block bytes |
| `StateMainnet_ReadSlot` | Read single field from raw state bytes |
| `StateMainnet_ReadValidatorCount` | Count validators from raw state bytes |
| `StateMainnet_ReadBalance` | Read one balance from raw state bytes |

## Running

```bash
# Verify library tests pass
cd dynamic-ssz && go test ./... -count=1 -timeout=5m

# Run benchmarks
cd .. && go test -run=^$ -bench=. -benchmem -count=3 -timeout=10m
```

## References

- [dynamic-ssz](https://github.com/pk910/dynamic-ssz)
- [SSZ Spec](https://github.com/ethereum/consensus-specs/blob/master/ssz/simple-serialize.md)
- [rkyv](https://github.com/rkyv/rkyv) — Rust zero-copy serialization
- [ssz-benchmark](https://github.com/pk910/ssz-benchmark) — Comparative SSZ benchmarks
