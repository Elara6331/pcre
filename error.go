package pcre

import (
	"fmt"
	"unsafe"

	"go.arsenm.dev/pcre/lib"

	"modernc.org/libc"
	"modernc.org/libc/sys/types"
)

var pce pcreError

// pcreError is meant to be manually allocated
// for when pcre requires a pointer to store an error
// such as for Xpcre2_compile_8().
type pcreError struct {
	errCode   int32
	errOffset lib.Tsize_t
}

// allocError manually allocates memory for the
// pcreError struct.
//
// libc.Xfree() should be called on the returned
// pointer once it is no longer needed.
func allocError(tls *libc.TLS) uintptr {
	return libc.Xmalloc(tls, types.Size_t(unsafe.Sizeof(pce)))
}

// addErrCodeOffset adds the offset of the error code
// within the pcreError struct to a given pointer
func addErrCodeOffset(p uintptr) uintptr {
	ptrOffset := unsafe.Offsetof(pce.errCode)
	return p + ptrOffset
}

// addErrOffsetOffset adds the offset of the error
// offset within the pcreError struct to a given pointer
func addErrOffsetOffset(p uintptr) uintptr {
	offsetOffset := unsafe.Offsetof(pce.errOffset)
	return p + offsetOffset
}

// ptrToError converts the given pointer to a Go error
func ptrToError(tls *libc.TLS, pe uintptr) *PcreError {
	eo := *(*pcreError)(unsafe.Pointer(pe))

	err := codeToError(tls, eo.errCode)
	err.offset = eo.errOffset

	return err
}

// codeToError converts the given error code into a Go error
func codeToError(tls *libc.TLS, code int32) *PcreError {
	errBuf := make([]byte, 256)
	cErrBuf := uintptr(unsafe.Pointer(&errBuf[0]))

	// Get the textual error message associated with the code,
	// and store it in errBuf.
	msgLen := lib.Xpcre2_get_error_message_8(tls, code, cErrBuf, 256)

	return &PcreError{0, string(errBuf[:msgLen])}
}

// PcreError represents errors returned
// by underlying pcre2 functions.
type PcreError struct {
	offset lib.Tsize_t
	errStr string
}

// Error returns the string within the error,
// prepending the offset if it exists.
func (pe *PcreError) Error() string {
	if pe.offset == 0 {
		return pe.errStr
	}
	return fmt.Sprintf("offset %d: %s", pe.offset, pe.errStr)
}
