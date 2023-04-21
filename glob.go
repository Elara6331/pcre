package pcre

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"go.elara.ws/pcre/lib"
	"modernc.org/libc"
)

// ConvertGlob converts the given glob into a
// pcre regular expression, and then returns
// the result.
func ConvertGlob(glob string) (string, error) {
	tls := libc.NewTLS()
	defer tls.Close()

	// Get C string from given glob
	cGlob, err := libc.CString(glob)
	if err != nil {
		return "", err
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
		return "", codeToError(tls, ret)
	}
	defer lib.Xpcre2_converted_pattern_free_8(tls, outPtr)

	// Get output as byte slice
	out := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), outLen)
	// Convert output to string
	// This copies the data, so it's safe for later use
	return string(out), nil
}

// CompileGlob is a convenience function that converts
// a glob to a pcre regular expression and then compiles
// it.
func CompileGlob(glob string) (*Regexp, error) {
	pattern, err := ConvertGlob(glob)
	if err != nil {
		return nil, err
	}
	// Compile converted glob and return results
	return Compile(pattern)
}

// Glob returns a list of matches for the given glob pattern.
// It returns nil if there was no match. If the glob contains
// "**", it will recurse through the directory, which may be
// extremely slow depending on which directory is being searched.
func Glob(glob string) ([]string, error) {
	// If glob is empty, return nil
	if glob == "" {
		return nil, nil
	}

	// If the glob is a file path, return the file
	_, err := os.Lstat(glob)
	if err == nil {
		return []string{glob}, nil
	}

	// If the glob has no glob characters, return nil
	if !hasGlobChars(glob) {
		return nil, nil
	}

	// Split glob by filepath separator
	paths := strings.Split(glob, string(filepath.Separator))

	var splitDir []string
	// For every path in split list
	for _, path := range paths {
		// If glob characters forund, stop
		if hasGlobChars(path) {
			break
		}
		// Add path to splitDir
		splitDir = append(splitDir, path)
	}

	// Join splitDir and add filepath separator. This is the directory that will be searched.
	dir := filepath.Join(splitDir...)
	
	if filepath.IsAbs(glob) {
		dir = string(filepath.Separator) + dir
	}

	// If the directory is not accessible, return error
	_, err = os.Lstat(dir)
	if err != nil {
		return nil, err
	}

	// Compile glob pattern
	r, err := CompileGlob(glob)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var matches []string
	// If glob contains "**" (starstar), walk recursively. Otherwise, only search dir.
	if strings.Contains(glob, "**") {
		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if r.MatchString(path) {
				matches = append(matches, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			// Get full path of file
			path := filepath.Join(dir, file.Name())
			if r.MatchString(path) {
				matches = append(matches, path)
			}
		}
	}

	return matches, nil
}

// hasGlobChars checks if the string has any
// characters that are part of a glob.
func hasGlobChars(s string) bool {
	return strings.ContainsAny(s, "*[]?")
}
