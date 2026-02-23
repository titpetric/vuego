package vuego

import (
	"io/fs"
	"sort"
)

// OverlayFS overlays two filesystems.
//
// This allows extension of the lower filesystem with modified files,
// new files and encourages composition of the contents of a `fs.FS`.
type OverlayFS struct {
	chainFS []fs.FS
}

// NewOverlayFS will create a new *OverlayFS.
func NewOverlayFS(upper fs.FS, lower ...fs.FS) *OverlayFS {
	chainFS := append([]fs.FS{upper}, lower...)
	return &OverlayFS{
		chainFS: chainFS,
	}
}

// Open opens a file in the overlaid filesystem.
func (o *OverlayFS) Open(name string) (fs.File, error) {
	for _, chainfs := range o.chainFS {
		if chainfs == nil {
			continue
		}

		f, err := chainfs.Open(name)
		if err == nil {
			return f, nil
		}
	}
	return nil, fs.ErrNotExist
}

// ReadDir implements combined FS reading.
func (o *OverlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	merged := make(map[string]fs.DirEntry)
	var lastErr error

	// Iterate through chain (upper layers first) so upper layers override lower
	for _, chainfs := range o.chainFS {
		if chainfs == nil {
			continue
		}

		entries, err := fs.ReadDir(chainfs, name)
		if err == nil {
			for _, e := range entries {
				// Only add if not already present (upper layers take precedence)
				if _, exists := merged[e.Name()]; !exists {
					merged[e.Name()] = e
				}
			}
		} else {
			lastErr = err
		}
	}

	// If no filesystem had this directory, return error
	if len(merged) == 0 && lastErr != nil {
		return nil, lastErr
	}

	entries := make([]fs.DirEntry, 0, len(merged))
	for _, e := range merged {
		entries = append(entries, e)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

// Glob implements combined FS reading.
func (o *OverlayFS) Glob(pattern string) ([]string, error) {
	matchMap := make(map[string]struct{})

	for _, chainfs := range o.chainFS {
		if chainfs == nil {
			continue
		}

		matches, _ := fs.Glob(chainfs, pattern)
		for _, m := range matches {
			matchMap[m] = struct{}{}
		}
	}

	results := make([]string, 0, len(matchMap))
	for m := range matchMap {
		results = append(results, m)
	}

	sort.Strings(results)
	return results, nil
}
