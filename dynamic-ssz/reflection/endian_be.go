//go:build mips || mips64 || ppc64 || s390x

package reflection

// nativeEndianIsLittle is false on big-endian architectures.
// Bulk memory copies between SSZ buffers and Go integers are disabled
// because SSZ uses little-endian encoding which doesn't match native byte order.
const nativeEndianIsLittle = false
