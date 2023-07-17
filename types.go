package pcre

import "go.elara.ws/pcre/lib"

type CompileOption uint32

// Compile option bits
const (
	Anchored           = CompileOption(lib.DPCRE2_ANCHORED)
	AllowEmptyClass    = CompileOption(lib.DPCRE2_ALLOW_EMPTY_CLASS)
	AltBsux            = CompileOption(lib.DPCRE2_ALT_BSUX)
	AltCircumflex      = CompileOption(lib.DPCRE2_ALT_CIRCUMFLEX)
	AltVerbnames       = CompileOption(lib.DPCRE2_ALT_VERBNAMES)
	AutoCallout        = CompileOption(lib.DPCRE2_AUTO_CALLOUT)
	Caseless           = CompileOption(lib.DPCRE2_CASELESS)
	DollarEndOnly      = CompileOption(lib.DPCRE2_DOLLAR_ENDONLY)
	DotAll             = CompileOption(lib.DPCRE2_DOTALL)
	DupNames           = CompileOption(lib.DPCRE2_DUPNAMES)
	EndAnchored        = CompileOption(lib.DPCRE2_ENDANCHORED)
	Extended           = CompileOption(lib.DPCRE2_EXTENDED)
	FirstLine          = CompileOption(lib.DPCRE2_FIRSTLINE)
	Literal            = CompileOption(lib.DPCRE2_LITERAL)
	MatchInvalidUTF    = CompileOption(lib.DPCRE2_MATCH_INVALID_UTF)
	MactchUnsetBackref = CompileOption(lib.DPCRE2_MATCH_UNSET_BACKREF)
	Multiline          = CompileOption(lib.DPCRE2_MULTILINE)
	NeverBackslashC    = CompileOption(lib.DPCRE2_NEVER_BACKSLASH_C)
	NeverUCP           = CompileOption(lib.DPCRE2_NEVER_UCP)
	NeverUTF           = CompileOption(lib.DPCRE2_NEVER_UTF)
	NoAutoCapture      = CompileOption(lib.DPCRE2_NO_AUTO_CAPTURE)
	NoAutoPossess      = CompileOption(lib.DPCRE2_NO_AUTO_POSSESS)
	NoDotStarAnchor    = CompileOption(lib.DPCRE2_NO_DOTSTAR_ANCHOR)
	NoStartOptimize    = CompileOption(lib.DPCRE2_NO_START_OPTIMIZE)
	NoUTFCheck         = CompileOption(lib.DPCRE2_NO_UTF_CHECK)
	UCP                = CompileOption(lib.DPCRE2_UCP)
	Ungreedy           = CompileOption(lib.DPCRE2_UNGREEDY)
	UseOffsetLimit     = CompileOption(lib.DPCRE2_USE_OFFSET_LIMIT)
	UTF                = CompileOption(lib.DPCRE2_UTF)
)

type CalloutFlags uint32

const (
	CalloutStartMatch = CalloutFlags(lib.DPCRE2_CALLOUT_STARTMATCH)
	CalloutBacktrack  = CalloutFlags(lib.DPCRE2_CALLOUT_BACKTRACK)
)

// CalloutBlock contains the data passed to callout functions
type CalloutBlock struct {
	// Version contains the version number of the block format.
	// The current version is 2.
	Version uint32

	// CalloutNumber contains the number of the callout, in the range 0-255.
	// This is the number that follows "?C". For callouts with string arguments,
	// this will always be zero.
	CalloutNumber uint32

	// CaptureTop contains the number of the highest numbered substring
	// captured so far plus one. If no substrings have yet been captured,
	// CaptureTop will be set to 1.
	CaptureTop uint32

	// CaptureLast contains the number of the last substring that was captured.
	CaptureLast uint32

	// Substrings contains all of the substrings captured so far.
	Substrings []string

	Mark string

	// Subject contains the string passed to the match function.
	Subject string

	// StartMatch contains the offset within the subject at which the current match attempt started.
	StartMatch uint

	// CurrentPosition contains the offset of the current match pointer within the subject.
	CurrentPosition uint

	// PatternPosition contains the offset within the pattern string to the next item to be matched.
	PatternPosition uint

	// NextItemLength contains the length of the next item to be processed in the pattern string.
	NextItemLength uint

	// CalloutStringOffset contains the code unit offset to the start of the callout argument string within the original pattern string.
	CalloutStringOffset uint

	// CalloutString is the string for the callout. For numerical callouts, this will always be empty.
	CalloutString string

	// CalloutFlags contains the following flags:
	// 	CalloutStartMatch
	// This is set for the first callout after the start of matching for each new starting position in the subject.
	// 	CalloutBacktrack
	// This is set if there has been a matching backtrack since the previous callout, or since the start of matching if this is the first callout from a pcre2_match() run.
	//
	// Both bits are set when a backtrack has caused a "bumpalong" to a new starting position in the subject.
	CalloutFlags CalloutFlags
}
