package gitignore

import (
	"os"
	"testing"

	"github.com/git-lfs/wildmatch"
)

const ModeFile os.FileMode = 0
const ModeDir os.FileMode = os.ModeDir

func TestCreateExclude(t *testing.T) {
	t.Run("Simple pattern", func(t *testing.T) {
		pattern := "/simple-pattern"
		exclude := CreateExclude(pattern, "", 0)
		if exclude.Pattern != pattern {
			t.Errorf("pattern should be %s but is %s", pattern, exclude.Pattern)
		}
		if exclude.Flags != 0 {
			t.Errorf("pattern should have no flags")
		}
	})

	t.Run("Sets ExcFlagNegative", func(t *testing.T) {
		pattern := "!some/path"
		exclude := CreateExclude(pattern, "", 0)
		if exclude.Flags != ExcFlagNegative {
			t.Errorf("pattern should have ExcFlagNegative flag but has %s", exclude.Flags)
		}
	})

	t.Run("Sets ExcFlagMustbedir", func(t *testing.T) {
		pattern := "must/be/dir/"
		exclude := CreateExclude(pattern, "", 0)
		if exclude.Flags != ExcFlagMustbedir {
			t.Errorf("pattern should have ExcFlagMustbedir flag but has %s", exclude.Flags)
		}
	})

	t.Run("Sets ExcFlagEndswith", func(t *testing.T) {
		pattern := "*ends-with-this"
		exclude := CreateExclude(pattern, "", 0)
		expected := ExcFlagEndswith | ExcFlagNodir
		if exclude.Flags != expected {
			t.Errorf("pattern should have %s flags but has %s", expected, exclude.Flags)
		}
	})
}

func TestIsExcludedFromList(t *testing.T) {
	t.Run("basics", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("ignore-whole-dir/", "", 0)
		el.AddExclude("ignore-children-in-dir/*", "", 0)
		el.AddExclude("!ignore-children-in-dir/not-me", "", 0)

		AssertIsExcludedFromList(t, "ignore-whole-dir", ModeDir, Excluded, el)
		AssertIsExcludedFromList(t, "ignore-children-in-dir/not-me", ModeDir, NotExcluded, el)
	})

	t.Run("whitelisting", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("/*", "", 0)
		el.AddExclude("!not-excluded-dir/", "", 0)
		el.AddExclude("!not-excluded-file", "", 0)

		AssertIsExcludedFromList(t, "some_dir", ModeDir, Excluded, el)
		AssertIsExcludedFromList(t, "not-excluded-dir", ModeDir, NotExcluded, el)
		AssertIsExcludedFromList(t, "not-excluded-file", ModeFile, NotExcluded, el)

	})

	t.Run("With ExcFlagEndswith", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("*literal", "", 0)
		el.AddExclude("*.ignored", "", 0)
		el.AddExclude("*.*.too", "", 0)

		AssertIsExcludedFromList(t, "something-literal", ModeDir, Excluded, el)
		AssertIsExcludedFromList(t, "something-literaly-different", ModeDir, ExclusionUndecided, el)
		AssertIsExcludedFromList(t, "i-am.ignored", ModeFile, Excluded, el)
		AssertIsExcludedFromList(t, "i-am.ignored.too", ModeFile, Excluded, el)

	})

	t.Run("Test basename matchs", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me", "in/subfolder/", 0)

		AssertIsExcludedFromList(t, "in/subfolder/exclude-me", ModeFile, Excluded, el)
	})

	t.Run("removing / prefix from name", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me/too", "in/subfolder/", 0)

		AssertIsExcludedFromList(t, "in/subfolder/exclude-me/too", ModeFile, Excluded, el)
		// AssertIsExcludedFromList(t, "in/subfolder/exclude-me-not", false, ExclusionUndecided, el)
	})

	t.Run("pathname is shorter than exclude.Base", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me", "in/subfolder/", 0)

		// should bail early due to pathlen < baselen+1
		AssertIsExcludedFromList(t, "in/subfolde", ModeFile, ExclusionUndecided, el)
	})

	t.Run("pathname[baselen] is not /", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me", "in/subfolder/", 0)

		// should bail early due to pathname[baselen] != '/'
		AssertIsExcludedFromList(t, "in/subfolde", ModeFile, ExclusionUndecided, el)
	})

	t.Run("basename does not match", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me/too", "in/subfolder/", 0)

		// should bail early due to !fspathncmp(dirent.pathname, exclude.Base, baselen)
		AssertIsExcludedFromList(t, "in/subfolden", ModeFile, ExclusionUndecided, el)
	})

	t.Run("pattern prefix does not match name", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me/too", "in/subfolder/", 0)

		// should bail early due to !fspathncmp(pattern, name, prefix)
		AssertIsExcludedFromList(t, "in/subfolder/xxxxxxx-me/too", ModeFile, ExclusionUndecided, el)
	})

	t.Run("must be dir when pattern ends with /", func(t *testing.T) {
		el := CreateExcludeList()
		el.AddExclude("exclude-me/", "in/subfolder/", 0)

		// should bail early due to !dirent.modeType.IsDir()
		AssertIsExcludedFromList(t, "in/subfolder/exclude-me", ModeFile, ExclusionUndecided, el)
		AssertIsExcludedFromList(t, "in/subfolder/exclude-me", ModeDir, Excluded, el)
	})

	t.Run("consider wildmatch.SystemCase", func(t *testing.T) {
		previousCase := wildmatch.SystemCase
		wildmatch.SystemCase = nil
		el := CreateExcludeList()
		el.AddExclude("exclude-Dir/", "in/subfolder/", 0)
		el.AddExclude("exclude-file", "in/subfolder/", 0)

		AssertIsExcludedFromList(t, "in/subfolder/exclude-Dir", ModeDir, Excluded, el)
		AssertIsExcludedFromList(t, "in/subfolder/exclude-dir", ModeDir, ExclusionUndecided, el)
		AssertIsExcludedFromList(t, "in/subfolder/exclude-file", ModeFile, Excluded, el)
		wildmatch.SystemCase = previousCase
	})
}

func AssertIsExcludedFromList(t *testing.T, pathname string, mode os.FileMode, expected IsExcluded, el *ExcludeList) {
	dirent := mockDirent(pathname, mode)
	if result := IsExcludedFromList(*dirent, el); result != expected {
		t.Errorf("expected IsExcludedFromList to return %s for path %s but got %s", expected, pathname, result)
	}
}

func TestSimpleLength(t *testing.T) {
	t.Run("with glob", func(t *testing.T) {
		if len := SimpleLength("abc?"); len != 3 {
			t.Errorf("expected 3 got %d", len)
		}
	})

	t.Run("without glob", func(t *testing.T) {
		if len := SimpleLength("abc?"); len != 3 {
			t.Errorf("expected 3 got %d", len)
		}
	})
}

func TestIsExcludedString(t *testing.T) {
	excluded := Excluded
	notExcluded := NotExcluded
	exclusionUndecided := ExclusionUndecided
	var unknownValue IsExcluded = 5
	if excluded.String() != "Excluded" {
		t.Errorf("expected Excluded but got %s", excluded.String())
	}
	if notExcluded.String() != "NotExcluded" {
		t.Errorf("expected NotExcluded but got %s", notExcluded.String())
	}
	if exclusionUndecided.String() != "ExclusionUndecided" {
		t.Errorf("expected ExclusionUndecided but got %s", exclusionUndecided.String())
	}
	if unknownValue.String() != "UnknownValue(5)" {
		t.Errorf("expected UnknownValue(5) but got %s", unknownValue.String())
	}

}

func mockDirent(pathname string, mode os.FileMode) *Dirent {
	return &Dirent{
		pathname: pathname,
		basename: "",
		modeType: mode,
	}
}
