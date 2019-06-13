package gitignore

import (
	"fmt"
	"strings"
)

// ExcludeFlags stores flags for a single exclude pattern
type ExcludeFlags uint

const (
	// ExcFlagNodir The pattern contains a '/', not including a possible final '/' that sets ExecFlagMustbedir
	ExcFlagNodir ExcludeFlags = 1 << iota // 1
	// The git code contains no flag with value 2
	_
	// ExcFlagEndswith The pattern is an "ends-with" pattern
	ExcFlagEndswith // 4
	// ExcFlagMustbedir The pattern should only match for directories (ends with `/`)
	ExcFlagMustbedir // 8
	// ExcFlagNegative The pattern is negated (`!` prefixed)
	ExcFlagNegative // 16
)

type Exclude struct {
	Pattern   string
	Base      string
	Flags     ExcludeFlags
	SourcePos int
}

func InRed(txt string) string {
	return fmt.Sprintf("\033[31m%s\033[39m", txt)
}

func (e Exclude) String() string {
	return fmt.Sprintf(InRed("Exclude{")+"Pattern: %s, Base: %s, Flags:%s SourcePos: %d"+InRed("}"), e.Pattern, e.Base, e.Flags, e.SourcePos)
}

func (f ExcludeFlags) String() string {
	flags := []string{}
	if f&ExcFlagNodir != 0 {
		flags = append(flags, "ExcFlagNodir")
	}

	if f&ExcFlagEndswith != 0 {
		flags = append(flags, "ExcFlagEndswith")
	}

	if f&ExcFlagMustbedir != 0 {
		flags = append(flags, "ExcFlagMustbedir")
	}

	if f&ExcFlagNegative != 0 {
		flags = append(flags, "ExcFlagNegative")
	}

	return strings.Join(flags, "|")
}

// IsGlobSpecial taken from https://github.com/git/git/blob/v2.22.0/t/helper/test-ctype.c#L38
func IsGlobSpecial(char rune) bool {
	return strings.Contains("*?[\\", string(char))
}

// NoWildcard https://github.com/git/git/blob/v2.22.0/dir.c#L559-L562
func NoWildcard(pattern string) bool {
	firstGlob := strings.IndexFunc(pattern, IsGlobSpecial)
	return firstGlob == -1
}

func SimpleLength(pattern string) int {
	firstGlob := strings.IndexFunc(pattern, IsGlobSpecial)
	if firstGlob == -1 {
		return len(pattern)
	}
	return firstGlob
}

// ParseExcludePattern https://github.com/git/git/blob/v2.22.0/dir.c#L564-L600
func ParseExcludePattern(pattern string) (string, ExcludeFlags) {
	var flags ExcludeFlags
	if strings.HasPrefix(pattern, "!") {
		flags |= ExcFlagNegative
		pattern = pattern[1:]
	}

	if strings.HasSuffix(pattern, "/") {
		flags |= ExcFlagMustbedir
		pattern = pattern[:len(pattern)-1]
	}

	firstSlash := strings.Index(pattern, "/")
	if firstSlash == -1 || firstSlash == (len(pattern)-1) {
		flags |= ExcFlagNodir
	}

	if strings.HasPrefix(pattern, "*") && NoWildcard(pattern[1:]) {
		flags |= ExcFlagEndswith
	}

	return pattern, flags
}

// CreateExclude https://github.com/git/git/blob/v2.22.0/dir.c#L610-L622
func CreateExclude(pattern string, base string, srcpos int) *Exclude {
	pattern, flags := ParseExcludePattern(pattern)
	return &Exclude{
		Pattern:   pattern,
		Base:      base,
		Flags:     flags,
		SourcePos: srcpos,
	}
}
