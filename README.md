# autosszsearch

Autonomous research to improve SSZ (Simple Serialize) performance in Go.

Inspired by [rkyv](https://github.com/rkyv/rkyv)'s zero-copy deserialization, this project explores whether similar techniques can dramatically reduce SSZ decode latency and memory allocation in Ethereum consensus layer applications.

## Background

SSZ is the serialization format used by Ethereum's consensus layer. Current Go implementations (fastssz, dynamic-ssz, etc.) perform traditional deserialization: parse the byte buffer, allocate Go structs, copy data into fields. For a mainnet BeaconState (~17MB), this means ~185K allocations and ~45MB of heap usage per decode.

**rkyv's insight**: design the serialized format so it IS the in-memory format. Access is O(1) — just reinterpret bytes.

**SSZ's opportunity**: SSZ already uses offset-based layout for variable-length data. Fixed fields are little-endian at known positions. The format is partially amenable to zero-copy access without any format changes.

## Structure

```
autosszsearch/
├── program.md          # Agent instructions (the experiment loop)
├── prepare.go          # Fixed benchmark harness (DO NOT MODIFY)
├── types.go            # Ethereum Deneb types (DO NOT MODIFY)
├── experiment.go       # The file the agent modifies (ALL experiments here)
├── go.mod              # Dependencies (fixed)
├── res/                # Test data (real Ethereum mainnet blocks/states)
│   ├── block-mainnet.ssz
│   ├── block-mainnet-meta.json
│   ├── state-mainnet.ssz
│   └── state-mainnet-meta.json
└── results.tsv         # Experiment log (untracked)
```

## Key metrics

| Metric | Description | Goal |
|--------|-------------|------|
| ns/op | Nanoseconds per operation | Lower is better (primary) |
| B/op | Bytes allocated per operation | Lower is better (secondary) |
| allocs/op | Number of heap allocations | Lower is better (tertiary) |

## Research directions

1. **Zero-copy views**: Wrap `[]byte`, provide typed accessors. No allocation.
2. **Lazy deserialization**: Only materialize fields on access.
3. **Buffer pooling**: Reuse allocations via `sync.Pool`.
4. **Unsafe pointer tricks**: Direct memory reinterpretation for aligned fields.
5. **Offset caching**: Pre-compute container offset tables.
6. **Hybrid access**: Zero-copy for reads, traditional unmarshal for mutation.

## Running

```bash
# Establish baseline
go test -run=^$ -bench=. -benchmem -count=3 -timeout=10m

# Run autonomously (agent mode)
# The agent follows program.md instructions
```

## References

- [SSZ Spec](https://github.com/ethereum/consensus-specs/blob/master/ssz/simple-serialize.md)
- [rkyv](https://github.com/rkyv/rkyv) — Rust zero-copy serialization
- [dynamic-ssz](https://github.com/pk910/dynamic-ssz) — Go SSZ with dynamic field sizes
- [ssz-benchmark](https://github.com/pk910/ssz-benchmark) — Comparative SSZ benchmarks
