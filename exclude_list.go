package gitignore

// ExcludeList holds a list of Exclude elements
// It can be used in the future to hold additional metadata such as
// where these Exclude patterns came from (filename, cmd line flags, ...)
type ExcludeList struct {
	Excludes []*Exclude
}

// CreateExcludeList creates an empty ExcludeList
func CreateExcludeList() *ExcludeList {
	return &ExcludeList{
		Excludes: []*Exclude{},
	}
}

// AddExclude https://github.com/git/git/blob/v2.22.0/dir.c#L602-L626
func (list *ExcludeList) AddExclude(pattern string, base string, srcpos int) {
	list.Excludes = append(list.Excludes, CreateExclude(pattern, base, srcpos))
}
