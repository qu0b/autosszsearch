# autosszsearch

Autonomous research to improve SSZ performance by modifying the dynamic-ssz library internals.

## Setup

To set up a new experiment, work with the user to:

1. **Agree on a run tag**: propose a tag based on today's date (e.g. `mar17`). The branch `autosszsearch/<tag>` must not already exist — this is a fresh run.
2. **Create the branch**: `git checkout -b autosszsearch/<tag>` from current master.
3. **Read the in-scope files**: Read these files for full context:
   - `README.md` — repository context and research directions.
   - `api.go` — thin API wrappers that benchmarks call. You may modify this to wire up new library APIs.
   - `bench_test.go` — fixed benchmark harness. Do not modify.
   - `prepare.go` — fixed test data loading. Do not modify.
   - `types.go` — fixed Ethereum Deneb types. Do not modify.
   - `dynamic-ssz/` — **the library you modify**. Read key files:
     - `dynssz.go` — main API entry point
     - `reflection/unmarshal.go` — reflection-based decoder
     - `reflection/marshal.go` — reflection-based encoder
     - `reflection/sszsize.go` — size calculation
     - `reflection/treeroot.go` — hash tree root computation
     - `sszutils/decoder_buffer.go` — buffer-based decoder primitives
     - `sszutils/encoder_buffer.go` — buffer-based encoder primitives
     - `codegen/gen_decoder.go` — code-generated decoder
     - `codegen/gen_encoder.go` — code-generated encoder
     - `hasher/` — incremental hasher (PR #133)
4. **Verify test data exists**: Check that `res/` contains `block-mainnet.ssz` and `state-mainnet.ssz`.
5. **Initialize results.tsv**: Create `results.tsv` with just the header row.
6. **Run the existing tests**: `cd dynamic-ssz && go test ./... -count=1 -timeout=5m` to verify the library is healthy before modifying it.
7. **Confirm and go**: Confirm setup looks good.

Once you get confirmation, kick off the experimentation.

## Experimentation

Each experiment modifies the **dynamic-ssz library** (`dynamic-ssz/` directory) and measures the impact via Go benchmarks. You launch benchmarks from the repo root: `go test -run=^$ -bench=. -benchmem -count=3 -timeout=10m > run.log 2>&1`.

**What you CAN do:**
- Modify any file in `dynamic-ssz/` — this is the library you're optimizing. All techniques are fair game: buffer pooling, zero-copy views, unsafe pointer tricks, allocation reduction, offset caching, lazy deserialization, new APIs, etc.
- Modify `api.go` — to wire up new library APIs to the benchmarks. But `api.go` must only contain calls into the library, NOT decoding/encoding logic itself.

**What you CANNOT do:**
- Modify `bench_test.go`, `prepare.go`, or `types.go`. These are fixed.
- Add new Go module dependencies. You can only use what's already in `go.mod`.
- Modify the benchmark harness. The `Benchmark*` functions are the ground truth metrics.
- Put decoding/encoding logic in `api.go`. All optimization must live in the library.

**The goal**: Improve dynamic-ssz library performance (lower ns/op, lower B/op, lower allocs/op) while maintaining correctness. All improvements must be generalizable — no hardcoded byte offsets for specific types. The library must remain a general-purpose SSZ encoder/decoder.

**Key metrics** (in priority order):
1. **ns/op** — lower is better (primary)
2. **B/op** — bytes allocated per operation (secondary)
3. **allocs/op** — heap allocations per operation (tertiary)

**Correctness is non-negotiable**: All benchmarks validate results via HTR comparison and round-trip fidelity. If an optimization breaks correctness, the benchmark will fail.

**Library tests must pass**: After each modification, verify with `cd dynamic-ssz && go test ./... -count=1 -timeout=5m`. If library tests fail, the experiment is invalid.

**Simplicity criterion**: All else being equal, simpler is better. A small improvement that adds ugly complexity is not worth it. The library must remain maintainable and upstreamable.

**The first run**: Your very first run should always be to establish the baseline, so you will run the benchmarks without any modifications.

## Output format

Standard Go benchmark output:

```
BenchmarkBlockMainnet_Unmarshal-16       20000     54070 ns/op    161816 B/op    1515 allocs/op
BenchmarkStateMainnet_Unmarshal-16         100  12345678 ns/op  45000000 B/op  185000 allocs/op
```

Extract metrics: `grep "^Benchmark" run.log`

## Logging results

Log results to `results.tsv` (tab-separated).

```
commit	block_unmarshal_ns	state_unmarshal_ns	block_htr_ns	state_htr_ns	status	description
```

1. git commit hash (short, 7 chars)
2. block unmarshal ns/op — use 0 for crashes
3. state unmarshal ns/op — use 0 for crashes
4. block HTR ns/op — use 0 for crashes
5. state HTR ns/op — use 0 for crashes
6. status: `keep`, `discard`, or `crash`
7. short text description of what this experiment tried

## The experiment loop

LOOP FOREVER:

1. Look at the git state: the current branch/commit we're on
2. Identify an optimization opportunity in the dynamic-ssz library
3. Modify library files in `dynamic-ssz/`
4. Optionally modify `api.go` to wire up new library APIs
5. Run library tests: `cd dynamic-ssz && go test ./... -count=1 -timeout=5m > test.log 2>&1`
6. If tests fail, fix or revert. Check: `grep -c "^ok\|^FAIL" test.log`
7. git commit (from repo root, so both api.go and dynamic-ssz/ changes are in one commit)
8. Run benchmarks: `go test -run=^$ -bench=. -benchmem -count=3 -timeout=10m > run.log 2>&1`
9. Read results: `grep "^Benchmark" run.log`
10. If grep is empty, the run crashed. Run `tail -n 50 run.log` to diagnose.
11. Record results in `results.tsv` (do not commit results.tsv)
12. If performance improved (lower ns/op on key benchmarks), keep the commit
13. If performance is equal or worse, `git reset --hard HEAD~1`

**Timeout**: Each benchmark should take ~2-5 minutes. Kill and discard if >10 minutes.

**Crashes**: Fix if trivial, skip and log "crash" if fundamentally broken.

**NEVER STOP**: Once the loop begins, do NOT pause to ask the human. Run indefinitely until manually stopped. If you run out of ideas, re-read the library code for new angles, study the SSZ wire format, try combining previous near-misses, try more radical approaches.

## Research directions

### Buffer & Allocation Reduction
- Pool decoder/encoder buffers via `sync.Pool`
- Bulk-allocate structs (one allocation for container + all sub-structs)
- Reduce slice header allocations for fixed-size arrays

### Zero-Copy Access Patterns
- Add view-type APIs to the library that wrap `[]byte` and provide typed accessors
- Return slices pointing into the original buffer instead of copying
- Lazy deserialization — only materialize fields on access

### Decoder Optimization
- Reduce reflection overhead in hot paths
- Inline common type dispatching
- Pre-compute field offset tables and cache them

### Hasher Optimization (PR #133 baseline)
- The incremental hasher from PR #133 is already included
- Look for further improvements: batch hashing, parallelism, cache-friendly access

### Code Generator Improvements
- Generate more efficient decode/encode routines
- Generate zero-copy view types alongside traditional marshal/unmarshal

### Unsafe Optimizations
- Direct memory reinterpretation for aligned fixed-size fields
- `unsafe.Slice` to create slices over existing buffer data
- Eliminate bounds checks in proven-safe hot loops
