//go:build 386 || amd64 || arm || arm64 || loong64 || mips64le || mipsle || ppc64le || riscv64 || wasm

package reflection

// nativeEndianIsLittle is true on little-endian architectures.
// Bulk memory copies between SSZ buffers and Go integers are only
// valid when the native byte order matches SSZ's little-endian encoding.
const nativeEndianIsLittle = true
