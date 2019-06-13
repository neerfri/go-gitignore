package gitignore

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/git-lfs/wildmatch"
)

// Dirent stores the name and file system mode type of discovered file system
// entries.
type Dirent struct {
	pathname string
	basename string
	modeType os.FileMode
}

type IsExcluded int

const (
	Excluded           IsExcluded = 1
	NotExcluded        IsExcluded = 0
	ExclusionUndecided IsExcluded = -1
)

func (val IsExcluded) String() string {
	switch val {
	case Excluded:
		return "Excluded"
	case NotExcluded:
		return "NotExcluded"
	case ExclusionUndecided:
		return "ExclusionUndecided"
	}
	return fmt.Sprintf("UnknownValue(%d)", val)
}

func fspathncmp(a string, b string, count int) bool {
	//TODO: add case-sensitive
	// ignoreCase := true
	if wildmatch.SystemCase == nil {
		return strings.Compare(a[:count], b[:count]) == 0
	}
	return strings.EqualFold(a[:count], b[:count])
}

// MatchBasename https://github.com/git/git/blob/v2.22.0/dir.c#L935-L957
func MatchBasename(dirent Dirent, exclude *Exclude) bool {
	pattern := exclude.Pattern
	patternlen := len(pattern)
	prefix := SimpleLength(pattern)
	basename := path.Base(dirent.pathname)
	basenamelen := len(basename)

	if prefix == patternlen {
		if patternlen == basenamelen && fspathncmp(pattern, basename, basenamelen) {
			return true
		}
	} else if exclude.Flags&ExcFlagEndswith != 0 {
		/* "*literal" matching against "fooliteral" */
		if patternlen-1 <= basenamelen && fspathncmp(pattern[1:], basename[basenamelen-(patternlen-1):], patternlen-1) {
			return true
		}
	} else {
		wildmatch := wildmatch.NewWildmatch(pattern, wildmatch.CaseFold)
		return wildmatch.Match(basename)
	}
	return false
}

// MatchPathname https://github.com/git/git/blob/v2.22.0/dir.c#L959-L1016
func MatchPathname(dirent Dirent, exclude *Exclude) bool {
	pattern := exclude.Pattern
	pathlen := len(dirent.pathname)
	baselen := len(exclude.Base)
	if strings.HasSuffix(exclude.Base, "/") {
		baselen--
	}
	fmt.Printf("MatchPathname: match %s with %s based in %s\n", dirent.pathname, pattern, exclude.Base)

	// https://github.com/git/git/blob/v2.22.0/dir.c#L967-L975
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
	}

	fmt.Printf("MatchPathname: [pathlen:%d baselen:%d dirent.pathname:%s exclude.Base:%s]\n", pathlen, baselen, dirent.pathname, exclude.Base)
	// https://github.com/git/git/blob/v2.22.0/dir.c#L977-L984
	if pathlen < baselen+1 ||
		(baselen != 0 && dirent.pathname[baselen] != '/') ||
		!fspathncmp(dirent.pathname, exclude.Base, baselen) {
		fmt.Printf("MatchPathname: false [pathlen:%d baselen:%d dirent.pathname:%s exclude.Base:%s]\n", pathlen, baselen, dirent.pathname, exclude.Base)
		return false
	}

	name := dirent.pathname[baselen:]
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	var namelen = len(name)

	prefix := SimpleLength(pattern)
	fmt.Printf("MatchPathname: [prefix:%d pattern:%s pattern[prefix:]:%s]\n", prefix, pattern, pattern[prefix:])
	if prefix > 0 {
		if prefix > namelen {
			fmt.Printf("MatchPathname: false [prefix:%d namelen:%d]\n", prefix, namelen)
			return false
		}

		if !fspathncmp(pattern, name, prefix) {
			fmt.Printf("MatchPathname: false [pattern:%s name:%s prefix:%d]\n", pattern, name, prefix)
			return false
		}

		pattern = pattern[prefix:]
		name = name[prefix:]

		/*
		 * If the whole pattern did not have a wildcard,
		 * then our prefix match is all we need; we
		 * do not need to call fnmatch at all.
		 */
		if len(pattern) == 0 && len(name) == 0 {
			return true
		}

	}

	fmt.Printf("MatchPathname: wildmatch(%s, %s)\n", pattern, dirent.pathname)
	wildmatch := wildmatch.NewWildmatch(pattern, wildmatch.CaseFold)
	return wildmatch.Match(dirent.pathname)
}

// LastExcludeMatchingFromList Scan the given exclude list in reverse to see whether pathname
// should be ignored.  The first match (i.e. the last on the list), if
// any, determines the fate.  Returns the exclude_list element which
// matched, or NULL for undecided.
func LastExcludeMatchingFromList(dirent Dirent, el *ExcludeList) *Exclude {
	for i := len(el.Excludes) - 1; i >= 0; i-- {
		e := el.Excludes[i]
		fmt.Printf("LastExcludeMatchingFromList: %s %s %s\n", dirent.pathname, e, dirent.modeType)
		if e.Flags&ExcFlagMustbedir != 0 && !dirent.modeType.IsDir() {
			fmt.Printf("LastExcludeMatchingFromList: bail1\n")
			continue
		}

		if e.Flags&ExcFlagNodir != 0 {
			if MatchBasename(dirent, e) {
				return e
			}
			fmt.Printf("LastExcludeMatchingFromList: bail2\n")
			continue
		}

		// https://github.com/git/git/blob/v2.22.0/dir.c#L1060
		if len(e.Base) != 0 && !strings.HasSuffix(e.Base, "/") {
			panic("assertion failed, Exclude.Base  is invalid")
		}

		if MatchPathname(dirent, e) {
			return e
		}

		fmt.Printf("LastExcludeMatchingFromList: no_match\n")
	}
	return nil
}

// IsExcludedFromList Scan the list and let the last match determine the fate.
// Return 1 for exclude, 0 for include and -1 for undecided.
// https://github.com/git/git/blob/v2.22.0/dir.c#L1071-L1085
func IsExcludedFromList(dirent Dirent, el *ExcludeList) IsExcluded {
	var exclude = LastExcludeMatchingFromList(dirent, el)
	if exclude == nil {
		return -1
	} else if exclude.Flags&ExcFlagNegative != 0 {
		return 0
	} else {
		return 1
	}
}
