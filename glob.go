package pcre

import (
	"unsafe"

	"go.arsenm.dev/pcre/lib"
	"modernc.org/libc"
)

// CompileGlob converts the given glob into a
// pcre regular expression, and then compiles it,
// returning the result.
func CompileGlob(glob string) (*Regexp, error) {
	tls := libc.NewTLS()
	defer tls.Close()

	// Get C string from given glob
	cGlob, err := libc.CString(glob)
	if err != nil {
		return nil, err
	}
	defer libc.Xfree(tls, cGlob)
	// Convert length to size_t
	cGlobLen := lib.Tsize_t(len(glob))

	// Create null pointer
	outPtr := uintptr(0)
	// Get pointer to pointer
	cOutPtr := uintptr(unsafe.Pointer(&outPtr))

	// Create 0 size_t
	outLen := lib.Tsize_t(0)
	// Get pointer to size_t
	cOutLen := uintptr(unsafe.Pointer(&outLen))

	// Convert glob to regular expression
	ret := lib.Xpcre2_pattern_convert_8(
		tls,
		cGlob,
		cGlobLen,
		lib.DPCRE2_CONVERT_GLOB,
		cOutPtr,
		cOutLen,
		0,
	)
	if ret != 0 {
		return nil, codeToError(tls, ret)
	}

	// Get output as byte slice
	out := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), outLen)
	// Convert output to string 
	// This copies the data, so it's safe for later use
	pattern := string(out)

	// Compile converted glob and return results
	return Compile(pattern)
}
