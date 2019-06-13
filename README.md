# go-gitignore

A go library to parse `.gitignore` patterns and match pattern lists against a path.

The code aims to be a close translation of the original implementation in the git codebase starting at [`is_excluded_from_list`](https://github.com/git/git/blob/6a6c0f10a70a6eb101c213b09ae82a9cad252743/dir.c#L1071-L1085) and [`add_excludes`](https://github.com/git/git/blob/6a6c0f10a70a6eb101c213b09ae82a9cad252743/dir.c#L769-L842)

### Prior Art

* https://github.com/zabawaba99/go-gitignore
* https://github.com/helm/helm/tree/v2.14.1/pkg/ignore
