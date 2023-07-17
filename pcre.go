// Package pcre is a library that provides pcre2 regular expressions
// in pure Go, allowing for features such as cross-compiling.
//
// The lib directory contains source code automatically translated from
// pcre2's C source code for each supported architecture and/or OS.
// This package wraps the automatically-translated source to provide a
// safe interface as close to Go's regexp library as possible.
package pcre

import (
	"os"
	"runtime"
	"strconv"
	"sync"
	"unsafe"

	"go.elara.ws/pcre/lib"

	"modernc.org/libc"
)

// Version returns the version of pcre2 embedded in this library.
func Version() string { return lib.DPACKAGE_VERSION }

// Regexp represents a pcre2 regular expression
type Regexp struct {
	mtx  *sync.Mutex
	expr string
	re   uintptr
	mctx uintptr
	tls  *libc.TLS

	calloutMtx *sync.Mutex
	callout    *func(tls *libc.TLS, cbptr, data uintptr) int32
}

// Compile runs CompileOpts with no options.
//
// Close() should be called on the returned expression
// once it is no longer needed.
func Compile(pattern string) (*Regexp, error) {
	return CompileOpts(pattern, 0)
}

// CompileOpts compiles the provided pattern using the given options.
//
// Close() should be called on the returned expression
// once it is no longer needed.
func CompileOpts(pattern string, options CompileOption) (*Regexp, error) {
	tls := libc.NewTLS()

	// Get C string of pattern
	cPattern, err := libc.CString(pattern)
	if err != nil {
		return nil, err
	}
	// Free the string when done
	defer libc.Xfree(tls, cPattern)

	// Allocate new error
	cErr := allocError(tls)
	// Free error when done
	defer libc.Xfree(tls, cErr)

	// Get error offsets
	errPtr := addErrCodeOffset(cErr)
	errOffsetPtr := addErrOffsetOffset(cErr)

	// Convert pattern length to size_t type
	cPatLen := lib.Tsize_t(len(pattern))

	// Compile expression
	r := lib.Xpcre2_compile_8(tls, cPattern, cPatLen, uint32(options), errPtr, errOffsetPtr, 0)
	if r == 0 {
		return nil, ptrToError(tls, cErr)
	}

	// Create regexp instance
	regex := Regexp{
		expr:       pattern,
		mtx:        &sync.Mutex{},
		re:         r,
		mctx:       lib.Xpcre2_match_context_create_8(tls, 0),
		tls:        tls,
		calloutMtx: &sync.Mutex{},
	}

	// Make sure resources are freed if GC collects the
	// regular expression.
	runtime.SetFinalizer(&regex, func(r *Regexp) error {
		return r.Close()
	})

	return &regex, nil
}

// MustCompile compiles the given pattern and panics
// if there was an error
//
// Close() should be called on the returned expression
// once it is no longer needed.
func MustCompile(pattern string) *Regexp {
	rgx, err := Compile(pattern)
	if err != nil {
		panic(err)
	}
	return rgx
}

// MustCompileOpts compiles the given pattern with the given
// options and panics if there was an error.
//
// Close() should be called on the returned expression
// once it is no longer needed.
func MustCompileOpts(pattern string, options CompileOption) *Regexp {
	rgx, err := CompileOpts(pattern, options)
	if err != nil {
		panic(err)
	}
	return rgx
}

// Find returns the leftmost match of the regular expression.
// A return value of nil indicates no match.
func (r *Regexp) Find(b []byte) []byte {
	matches, err := r.match(b, 0, false)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return nil
	}
	match := matches[0]
	return b[match[0]:match[1]]
}

// FindIndex returns a two-element slice of integers
// representing the location of the leftmost match of the
// regular expression.
func (r *Regexp) FindIndex(b []byte) []int {
	matches, err := r.match(b, 0, false)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return nil
	}
	match := matches[0]

	return []int{int(match[0]), int(match[1])}
}

// FindAll returns all matches of the regular expression.
// A return value of nil indicates no match.
func (r *Regexp) FindAll(b []byte, n int) [][]byte {
	matches, err := r.match(b, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 || n == 0 {
		return nil
	}
	if n > 0 && len(matches) > n {
		matches = matches[:n]
	}

	out := make([][]byte, len(matches))
	for index, match := range matches {
		out[index] = b[match[0]:match[1]]
	}

	return out
}

// FindAll returns indices of all matches of the
// regular expression. A return value of nil indicates
// no match.
func (r *Regexp) FindAllIndex(b []byte, n int) [][]int {
	matches, err := r.match(b, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 || n == 0 {
		return nil
	}
	if n > 0 && len(matches) > n {
		matches = matches[:n]
	}

	out := make([][]int, len(matches))
	for index, match := range matches {
		out[index] = []int{int(match[0]), int(match[1])}
	}
	return out
}

// FindSubmatch returns a slice containing the match as the
// first element, and the submatches as the subsequent elements.
func (r *Regexp) FindSubmatch(b []byte) [][]byte {
	matches, err := r.match(b, 0, false)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return nil
	}
	match := matches[0]

	out := make([][]byte, 0, len(match)/2)
	for i := 0; i < len(match); i += 2 {
		out = append(out, b[match[i]:match[i+1]])
	}
	return out
}

// FindSubmatchIndex returns a slice of index pairs representing
// the match and submatches, if any.
func (r *Regexp) FindSubmatchIndex(b []byte) []int {
	matches, err := r.match(b, 0, false)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return nil
	}
	match := matches[0]

	out := make([]int, len(match))
	for index, offset := range match {
		out[index] = int(offset)
	}

	return out
}

// FindAllSubmatch returns a slice of all matches and submatches
// of the regular expression. It will return no more than n matches.
// If n < 0, it will return all matches.
func (r *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	matches, err := r.match(b, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 || n == 0 {
		return nil
	}
	if n > 0 && len(matches) > n {
		matches = matches[:n]
	}

	out := make([][][]byte, len(matches))
	for index, match := range matches {
		outMatch := make([][]byte, 0, len(match)/2)

		for i := 0; i < len(match); i += 2 {
			outMatch = append(outMatch, b[match[i]:match[i+1]])
		}

		out[index] = outMatch
	}

	return out
}

// FindAllSubmatch returns a slice of all indeces representing the
// locations of matches and submatches, if any, of the regular expression.
// It will return no more than n matches. If n < 0, it will return all matches.
func (r *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	matches, err := r.match(b, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 || n == 0 {
		return nil
	}
	if n > 0 && len(matches) > n {
		matches = matches[:n]
	}

	out := make([][]int, len(matches))
	for index, match := range matches {
		offsets := make([]int, len(match))

		for index, offset := range match {
			offsets[index] = int(offset)
		}

		out[index] = offsets
	}

	return out
}

// FindString is the String version of Find
func (r *Regexp) FindString(s string) string {
	return string(r.Find([]byte(s)))
}

// FindStringIndex is the String version of FindIndex
func (r *Regexp) FindStringIndex(s string) []int {
	return r.FindIndex([]byte(s))
}

// FinAllString is the String version of FindAll
func (r *Regexp) FindAllString(s string, n int) []string {
	matches := r.FindAll([]byte(s), n)

	out := make([]string, len(matches))
	for index, match := range matches {
		out[index] = string(match)
	}
	return out
}

// FindAllStringIndex is the String version of FindIndex
func (r *Regexp) FindAllStringIndex(s string, n int) [][]int {
	return r.FindAllIndex([]byte(s), n)
}

// FindStringSubmatch is the string version of FindSubmatch
func (r *Regexp) FindStringSubmatch(s string) []string {
	matches := r.FindSubmatch([]byte(s))

	out := make([]string, len(matches))
	for index, match := range matches {
		out[index] = string(match)
	}
	return out
}

// FindStringSubmatchIndex is the String version of FindSubmatchIndex
func (r *Regexp) FindStringSubmatchIndex(s string) []int {
	return r.FindSubmatchIndex([]byte(s))
}

// FindAllStringSubmatch is the String version of FindAllSubmatch
func (r *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	matches := r.FindAllSubmatch([]byte(s), n)

	out := make([][]string, len(matches))
	for index, match := range matches {
		outMatch := make([]string, len(match))

		for index, byteMatch := range match {
			outMatch[index] = string(byteMatch)
		}

		out[index] = outMatch
	}

	return out
}

// FindAllStringSubmatchIndex is the String version of FindAllSubmatchIndex
func (r *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	return r.FindAllSubmatchIndex([]byte(s), n)
}

// Match reports whether b contains a match of the regular expression
func (r *Regexp) Match(b []byte) bool {
	return r.Find(b) != nil
}

// MatchString is the String version of Match
func (r *Regexp) MatchString(s string) bool {
	return r.Find([]byte(s)) != nil
}

// NumSubexp returns the number of parenthesized subexpressions
// in the regular expression.
func (r *Regexp) NumSubexp() int {
	return int(r.patternInfo(lib.DPCRE2_INFO_CAPTURECOUNT))
}

// ReplaceAll returns a copy of src, replacing matches of the
// regular expression with the replacement text repl.
// Inside repl, $ signs are interpreted as in Expand,
// so for instance $1 represents the text of the first
// submatch and $name would represent the text of the
// subexpression called "name".
func (r *Regexp) ReplaceAll(src, repl []byte) []byte {
	matches, err := r.match(src, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return src
	}

	out := make([]byte, len(src))
	copy(out, src)

	var diff int64
	for _, match := range matches {
		replStr := os.Expand(string(repl), func(s string) string {
			i, err := strconv.Atoi(s)
			if err != nil {
				i = r.SubexpIndex(s)
				if i == -1 {
					return ""
				}
			}

			// If there given match does not exist, return empty string
			if i == 0 || len(match) < (2*i)+1 {
				return ""
			}

			// Return match
			return string(src[match[2*i]:match[(2*i)+1]])
		})
		// Replace replacement string with expanded string
		repl := []byte(replStr)

		// Replace bytes with new replacement string
		diff, out = replaceBytes(out, repl, match[0], match[1], diff)
	}

	return out
}

// ReplaceAllFunc returns a copy of src in which all matches of the
// regular expression have been replaced by the return value of function
// repl applied to the matched byte slice. The replacement returned by
// repl is substituted directly, without using Expand.
func (r *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	matches, err := r.match(src, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return src
	}

	out := make([]byte, len(src))
	copy(out, src)

	var diff int64
	for _, match := range matches {
		replBytes := repl(src[match[0]:match[1]])
		diff, out = replaceBytes(out, replBytes, match[0], match[1], diff)
	}

	return out
}

// ReplaceAllLiteral returns a copy of src, replacing matches of
// the regular expression with the replacement bytes repl.
// The replacement is substituted directly, without using Expand.
func (r *Regexp) ReplaceAllLiteral(src, repl []byte) []byte {
	matches, err := r.match(src, 0, true)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		return src
	}

	out := make([]byte, len(src))
	copy(out, src)

	var diff int64
	for _, match := range matches {
		diff, out = replaceBytes(out, repl, match[0], match[1], diff)
	}

	return out
}

// ReplaceAllString is the String version of ReplaceAll
func (r *Regexp) ReplaceAllString(src, repl string) string {
	return string(r.ReplaceAll([]byte(src), []byte(repl)))
}

// ReplaceAllStringFunc is the String version of ReplaceAllFunc
func (r *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	return string(r.ReplaceAllFunc([]byte(src), func(b []byte) []byte {
		return []byte(repl(string(b)))
	}))
}

// ReplaceAllLiteralString is the String version of ReplaceAllLiteral
func (r *Regexp) ReplaceAllLiteralString(src, repl string) string {
	return string(r.ReplaceAllLiteral([]byte(src), []byte(repl)))
}

// Split slices s into substrings separated by the
// expression and returns a slice of the substrings
// between those expression matches.
//
// Example:
//
//	s := regexp.MustCompile("a*").Split("abaabaccadaaae", 5)
//	// s: ["", "b", "b", "c", "cadaaae"]
//
// The count determines the number of substrings to return:
//
//	n > 0: at most n substrings; the last substring will be the unsplit remainder.
//	n == 0: the result is nil (zero substrings)
//	n < 0: all substrings
func (r *Regexp) Split(s string, n int) []string {
	if n == 0 {
		return nil
	}

	if len(r.expr) > 0 && len(s) == 0 {
		return []string{""}
	}

	matches := r.FindAllStringIndex(s, n)
	strings := make([]string, 0, len(matches))

	beg := 0
	end := 0
	for _, match := range matches {
		if n > 0 && len(strings) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			strings = append(strings, s[beg:end])
		}
		beg = match[1]
	}

	if end != len(s) {
		strings = append(strings, s[beg:])
	}

	return strings
}

// String returns the text of the regular expression
// used for compilation.
func (r *Regexp) String() string {
	return r.expr
}

// SubexpIndex returns the index of the subexpression
// with the given name, or -1 if there is no subexpression
// with that name.
func (r *Regexp) SubexpIndex(name string) int {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Get C string of name
	cName, err := libc.CString(name)
	if err != nil {
		panic(err)
	}

	// Get substring index from name
	ret := lib.Xpcre2_substring_number_from_name_8(r.tls, r.re, cName)

	// If no substring error returned, return -1.
	// If a different error is returned, panic.
	if ret == lib.DPCRE2_ERROR_NOSUBSTRING {
		return -1
	} else if ret < 0 {
		panic(codeToError(r.tls, ret))
	}

	// Return the index of the subexpression
	return int(ret)
}

// SetCallout sets a callout function that will be called at specified points in the matching operation.
// fn should return zero if it ran successfully or a non-zero integer to force an error.
// See https://www.pcre.org/current/doc/html/pcre2callout.html for more information.
func (r *Regexp) SetCallout(fn func(cb *CalloutBlock) int32) error {
	cfn := func(tls *libc.TLS, cbptr, data uintptr) int32 {
		ccb := (*lib.Tpcre2_callout_block_8)(unsafe.Pointer(cbptr))

		cb := &CalloutBlock{
			Version:             ccb.Fversion,
			CalloutNumber:       ccb.Fcallout_number,
			CaptureTop:          ccb.Fcapture_top,
			CaptureLast:         ccb.Fcapture_last,
			Mark:                libc.GoString(ccb.Fmark),
			StartMatch:          uint(ccb.Fstart_match),
			CurrentPosition:     uint(ccb.Fcurrent_position),
			PatternPosition:     uint(ccb.Fpattern_position),
			NextItemLength:      uint(ccb.Fnext_item_length),
			CalloutStringOffset: uint(ccb.Fcallout_string_offset),
			CalloutFlags:        CalloutFlags(ccb.Fcallout_flags),
		}

		subjectBytes := unsafe.Slice((*byte)(unsafe.Pointer(ccb.Fsubject)), ccb.Fsubject_length)
		cb.Subject = string(subjectBytes)

		calloutStrBytes := unsafe.Slice((*byte)(unsafe.Pointer(ccb.Fcallout_string)), ccb.Fcallout_string_length)
		cb.CalloutString = string(calloutStrBytes)

		ovecSlice := unsafe.Slice((*lib.Tsize_t)(unsafe.Pointer(ccb.Foffset_vector)), (ccb.Fcapture_top*2)-1)[2:]
		for i := 0; i < len(ovecSlice); i += 2 {
			if i+1 >= len(ovecSlice) {
				cb.Substrings = append(cb.Substrings, cb.Subject[ovecSlice[i]:])
			} else {
				cb.Substrings = append(cb.Substrings, cb.Subject[ovecSlice[i]:ovecSlice[i+1]])
			}
		}

		return fn(cb)
	}

	r.calloutMtx.Lock()
	defer r.calloutMtx.Unlock()

	// Prevent callout function from being GC'd
	r.callout = &cfn

	ret := lib.Xpcre2_set_callout_8(r.tls, r.mctx, *(*uintptr)(unsafe.Pointer(&cfn)), 0)
	if ret < 0 {
		return codeToError(r.tls, ret)
	}
	return nil
}

// replaceBytes replaces the bytes at a given location, and returns a new
// offset, based on how much bigger or smaller the slice got after replacement
func replaceBytes(src, repl []byte, sOff, eOff lib.Tsize_t, diff int64) (int64, []byte) {
	var out []byte
	out = append(
		src[:int64(sOff)+diff],
		append(
			repl,
			src[int64(eOff)+diff:]...,
		)...,
	)

	return diff + int64(len(out)-len(src)), out
}

// match calls the underlying pcre match functions. It re-runs the functions
// until no matches are found if multi is set to true.
func (r *Regexp) match(b []byte, options uint32, multi bool) ([][]lib.Tsize_t, error) {
	if len(b) == 0 {
		return nil, nil
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Create a C pointer to the subject
	sp := unsafe.Pointer(&b[0])
	cSubject := uintptr(sp)
	// Convert the size of the subject to a C size_t type
	cSubjectLen := lib.Tsize_t(len(b))

	// Create match data using the pattern to figure out the buffer size
	md := lib.Xpcre2_match_data_create_from_pattern_8(r.tls, r.re, 0)
	if md == 0 {
		panic("error creating match data")
	}
	// Free the match data at the end of the function
	defer lib.Xpcre2_match_data_free_8(r.tls, md)

	var offset lib.Tsize_t
	var out [][]lib.Tsize_t
	// While the offset is less than the length of the subject
	for offset < cSubjectLen {
		// Execute expression on subject
		ret := lib.Xpcre2_match_8(r.tls, r.re, cSubject, cSubjectLen, offset, options, md, r.mctx)
		if ret < 0 {
			// If no match found, break
			if ret == lib.DPCRE2_ERROR_NOMATCH {
				break
			}

			return nil, codeToError(r.tls, ret)
		} else {
			// Get amount of pairs in output vector
			pairAmt := lib.Xpcre2_get_ovector_count_8(r.tls, md)
			// Get pointer to output vector
			ovec := lib.Xpcre2_get_ovector_pointer_8(r.tls, md)
			// Create a Go slice using the output vector as the underlying array
			slice := unsafe.Slice((*lib.Tsize_t)(unsafe.Pointer(ovec)), pairAmt*2)

			// Create a new slice and copy the elements from the slice
			// This is required because the match data will be freed in
			// a defer, and that would cause a panic every time the slice
			// is used later.
			matches := make([]lib.Tsize_t, len(slice))
			copy(matches, slice)

			// If the two indices are the same (empty string), and the match is not
			// immediately after another match, add it to the output and increment the
			// offset. Otherwise, increment the offset and ignore the match.
			if slice[0] == slice[1] && len(out) > 0 && slice[0] != out[len(out)-1][1] {
				out = append(out, matches)
				offset = slice[1] + 1
				continue
			} else if slice[0] == slice[1] {
				offset = slice[1] + 1
				continue
			}

			// Add the match to the output
			out = append(out, matches)
			// Set the next offset to the end index of the match
			offset = matches[1]
		}

		// If multiple matches disabled, break
		if !multi {
			break
		}
	}
	return out, nil
}

// patternInfo calls the underlying pcre pattern info function
// and returns information about the compiled regular expression
func (r *Regexp) patternInfo(what uint32) (out uint32) {
	// Create a C pointer to the output integer
	cOut := uintptr(unsafe.Pointer(&out))
	// Get information about the compiled pattern
	lib.Xpcre2_pattern_info_8(r.tls, r.re, what, cOut)
	return
}

// Close frees resources used by the regular expression.
func (r *Regexp) Close() error {
	if r == nil {
		return nil
	}

	// Close thread-local storage
	defer r.tls.Close()

	// Free the compiled code
	lib.Xpcre2_code_free_8(r.tls, r.re)
	// Free the match context
	lib.Xpcre2_match_context_free_8(r.tls, r.mctx)
	// Set regular expression to null
	r.re = 0

	return nil
}
