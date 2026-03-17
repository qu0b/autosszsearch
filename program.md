# autosszsearch

Autonomous research to improve SSZ (Simple Serialize) performance in Go, inspired by zero-copy techniques from Rust's rkyv library.

## Setup

To set up a new experiment, work with the user to:

1. **Agree on a run tag**: propose a tag based on today's date (e.g. `mar17`). The branch `autosszsearch/<tag>` must not already exist — this is a fresh run.
2. **Create the branch**: `git checkout -b autosszsearch/<tag>` from current master.
3. **Read the in-scope files**: The repo is small. Read these files for full context:
   - `README.md` — repository context and research directions.
   - `prepare.go` — fixed constants, test data loading, benchmark harness. Do not modify.
   - `types.go` — Ethereum consensus types (Deneb fork). Do not modify.
   - `experiment.go` — the file you modify. SSZ optimization implementations.
4. **Verify test data exists**: Check that `res/` contains `block-mainnet.ssz` and `state-mainnet.ssz`. If not, tell the human to run `go run ./cmd/prepare`.
5. **Initialize results.tsv**: Create `results.tsv` with just the header row. The baseline will be recorded after the first run.
6. **Confirm and go**: Confirm setup looks good.

Once you get confirmation, kick off the experimentation.

## Experimentation

Each experiment is a Go benchmark suite. The benchmark script runs for a **fixed iteration count** and measures ns/op, B/op, and allocs/op. You launch it simply as: `go test -run=^$ -bench=. -benchmem -count=3 -timeout=10m > run.log 2>&1`.

**What you CAN do:**
- Modify `experiment.go` — this is the only file you edit. Everything is fair game: zero-copy views, lazy deserialization, buffer pooling, unsafe pointer tricks, SIMD-friendly layouts, custom allocators, offset caching, mmap-compatible access patterns, etc.

**What you CANNOT do:**
- Modify `prepare.go`. It is read-only. It contains the fixed benchmark harness, test data loading, and correctness validation.
- Modify `types.go`. The Ethereum types are fixed — they represent real-world usage.
- Install new Go modules. You can only use what's already in `go.mod`.
- Modify the benchmark harness. The `Benchmark*` functions in `prepare.go` are the ground truth metrics.

**The goal is simple: get the lowest ns/op for unmarshal operations while maintaining zero B/op where possible.** The key metrics are:
1. **ns/op** — lower is better (primary metric)
2. **B/op** — lower is better (secondary metric, measures allocations)
3. **allocs/op** — lower is better (tertiary metric)

Since the benchmark count is fixed, you don't need to worry about statistical noise — each run uses `-count=3` for stability. Everything is fair game: change the approach, use unsafe, pool buffers, use zero-copy views, lazy access, etc. The only constraint is that the code compiles and produces correct results (verified via HTR comparison).

**Correctness criterion**: All benchmark functions validate results against pre-computed Hash Tree Roots (HTRs). If an optimization breaks correctness, the benchmark will fail. Correctness is non-negotiable.

**Simplicity criterion**: All else being equal, simpler is better. A small improvement that adds ugly complexity is not worth it. Conversely, removing something and getting equal or better results is a great outcome. When evaluating whether to keep a change, weigh the complexity cost against the improvement magnitude.

**The first run**: Your very first run should always be to establish the baseline, so you will run the benchmark as is.

## Output format

The benchmark output looks like standard Go benchmark output:

```
BenchmarkBlockMainnet_Unmarshal-16       20000     54070 ns/op    161816 B/op    1515 allocs/op
BenchmarkStateMainnet_Unmarshal-16         100  12345678 ns/op  45000000 B/op  185000 allocs/op
BenchmarkBlockMainnet_ViewAccess-16     500000      2100 ns/op         0 B/op       0 allocs/op
```

You can extract the key metrics from the log file:

```
grep "^Benchmark" run.log
```

## Logging results

When an experiment is done, log it to `results.tsv` (tab-separated, NOT comma-separated — commas break in descriptions).

The TSV has a header row and 6 columns:

```
commit	block_ns_op	state_ns_op	block_allocs	status	description
```

1. git commit hash (short, 7 chars)
2. block unmarshal ns/op (e.g. 54070) — use 0 for crashes
3. state unmarshal ns/op (e.g. 12345678) — use 0 for crashes
4. block unmarshal allocs/op (e.g. 1515) — use 0 for crashes
5. status: `keep`, `discard`, or `crash`
6. short text description of what this experiment tried

Example:

```
commit	block_ns_op	state_ns_op	block_allocs	status	description
a1b2c3d	54070	12345678	1515	keep	baseline (standard unmarshal)
b2c3d4e	2100	500000	0	keep	zero-copy view with lazy field access
c3d4e5f	0	0	0	crash	unsafe pointer cast OOB panic
```

## The experiment loop

The experiment runs on a dedicated branch (e.g. `autosszsearch/mar17`).

LOOP FOREVER:

1. Look at the git state: the current branch/commit we're on
2. Tune `experiment.go` with an experimental idea by directly hacking the code.
3. git commit
4. Run the experiment: `go test -run=^$ -bench=. -benchmem -count=3 -timeout=10m > run.log 2>&1` (redirect everything — do NOT use tee or let output flood your context)
5. Read out the results: `grep "^Benchmark" run.log`
6. If the grep output is empty, the run crashed. Run `tail -n 50 run.log` to read the Go stack trace and attempt a fix. If you can't get things to work after more than a few attempts, give up.
7. Record the results in the tsv (NOTE: do not commit the results.tsv file, leave it untracked by git)
8. If the primary metric improved (lower ns/op), you "advance" the branch, keeping the git commit
9. If the primary metric is equal or worse, you git reset back to where you started

The idea is that you are a completely autonomous researcher trying things out. If they work, keep. If they don't, discard. And you're advancing the branch so that you can iterate. If you feel like you're getting stuck in some way, you can rewind but you should probably do this very very sparingly (if ever).

**Timeout**: Each benchmark run should take ~2-5 minutes total. If a run exceeds 10 minutes, kill it and treat it as a failure (discard and revert).

**Crashes**: If a run crashes (panic, or a bug, or etc.), use your judgment: If it's something dumb and easy to fix (e.g. a typo, a missing import), fix it and re-run. If the idea itself is fundamentally broken, just skip it, log "crash" as the status in the tsv, and move on.

**NEVER STOP**: Once the experiment loop has begun (after the initial setup), do NOT pause to ask the human if you should continue. Do NOT ask "should I keep going?" or "is this a good stopping point?". The human might be asleep, or gone from a computer and expects you to continue working *indefinitely* until you are manually stopped. You are autonomous. If you run out of ideas, think harder — re-read the research directions in README.md, study the SSZ wire format, try combining previous near-misses, try more radical approaches. The loop runs until the human interrupts you, period.

## Research directions

When looking for ideas, consider these SSZ optimization approaches:

### Zero-Copy Views (rkyv-inspired)
- Wrap raw `[]byte` buffer, provide accessor methods that read directly
- O(1) "deserialization" — just store a reference to the buffer
- O(1) field access — compute offset, read bytes from buffer
- Fixed fields: `binary.LittleEndian.Uint64(buf[offset:])` — no allocation
- Variable fields: read 4-byte offset, return sub-view
- Nested containers: chain offset lookups

### Lazy Deserialization
- Only deserialize fields when accessed
- Cache deserialized values for repeated access
- Particularly valuable for BeaconState (17MB) when only a few fields are needed

### Buffer Pooling & Reuse
- `sync.Pool` for temporary buffers
- Pre-allocate common sizes
- Reuse decoder state across calls

### Unsafe Optimizations
- Direct pointer arithmetic for aligned fixed fields
- Slice header manipulation to avoid copies
- `unsafe.String` / `unsafe.Slice` for zero-copy byte access

### Offset Table Caching
- Pre-compute and cache offset tables for containers
- Amortize offset parsing across multiple field accesses

### SIMD-Friendly Layouts
- Align data for vectorized operations
- Batch decode multiple validators simultaneously

### Hybrid Approaches
- Zero-copy for read-only access, traditional unmarshal for mutation
- Selective materialization — only allocate what you need
