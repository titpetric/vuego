package vuego_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// entryNames returns the names of dir entries in the order given.
func entryNames(entries []fs.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// readAll reads a file opened from an overlay and returns its contents.
func readAll(t *testing.T, fsys fs.FS, name string) string {
	t.Helper()

	f, err := fsys.Open(name)
	assert.NoError(t, err)
	defer f.Close()

	data, err := fs.ReadFile(fsys, name)
	assert.NoError(t, err)
	return string(data)
}

func TestNewOverlayFS(t *testing.T) {
	t.Run("the first argument is the upper layer", func(t *testing.T) {
		overlay := vuego.NewOverlayFS(
			fstest.MapFS{"file.txt": {Data: []byte("upper")}},
			fstest.MapFS{"file.txt": {Data: []byte("lower")}},
		)

		assert.Equal(t, "upper", readAll(t, overlay, "file.txt"))
	})

	t.Run("an upper layer alone is a valid overlay", func(t *testing.T) {
		overlay := vuego.NewOverlayFS(fstest.MapFS{"file.txt": {Data: []byte("only")}})
		assert.Equal(t, "only", readAll(t, overlay, "file.txt"))
	})

	t.Run("satisfies fs.FS", func(t *testing.T) {
		var fsys fs.FS = vuego.NewOverlayFS(fstest.MapFS{})
		assert.NotNil(t, fsys)
	})
}

func TestOverlayFS_Open(t *testing.T) {
	tests := []struct {
		name  string
		chain []fs.FS
		open  string
		want  string
		err   bool
	}{
		{
			name:  "upper only",
			chain: []fs.FS{fstest.MapFS{"file.txt": {Data: []byte("upper content")}}, fstest.MapFS{}},
			open:  "file.txt",
			want:  "upper content",
		},
		{
			name:  "falls through to the lower layer",
			chain: []fs.FS{fstest.MapFS{}, fstest.MapFS{"file.txt": {Data: []byte("lower content")}}},
			open:  "file.txt",
			want:  "lower content",
		},
		{
			name: "upper overrides lower",
			chain: []fs.FS{
				fstest.MapFS{"file.txt": {Data: []byte("upper content")}},
				fstest.MapFS{"file.txt": {Data: []byte("lower content")}},
			},
			open: "file.txt",
			want: "upper content",
		},
		{
			name:  "nested path",
			chain: []fs.FS{fstest.MapFS{"dir/file.txt": {Data: []byte("upper")}}, fstest.MapFS{}},
			open:  "dir/file.txt",
			want:  "upper",
		},
		{
			name:  "a nil layer is skipped",
			chain: []fs.FS{fstest.MapFS{"file.txt": {Data: []byte("upper content")}}, nil},
			open:  "file.txt",
			want:  "upper content",
		},
		{
			name:  "searches the whole chain",
			chain: []fs.FS{fstest.MapFS{}, fstest.MapFS{}, fstest.MapFS{"file.txt": {Data: []byte("found in fs3")}}},
			open:  "file.txt",
			want:  "found in fs3",
		},
		{
			name:  "missing file",
			chain: []fs.FS{fstest.MapFS{}, fstest.MapFS{}},
			open:  "file.txt",
			err:   true,
		},
		{
			name:  "every layer nil",
			chain: []fs.FS{nil, nil, nil},
			open:  "file.txt",
			err:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			overlay := vuego.NewOverlayFS(test.chain[0], test.chain[1:]...)

			if test.err {
				_, err := overlay.Open(test.open)
				assert.Error(t, err)
				return
			}

			assert.Equal(t, test.want, readAll(t, overlay, test.open))
		})
	}
}

func TestOverlayFS_ReadDir(t *testing.T) {
	tests := []struct {
		name  string
		chain []fs.FS
		dir   string
		want  []string
		err   bool
	}{
		{
			name: "upper only",
			chain: []fs.FS{fstest.MapFS{
				"file1.txt": {Data: []byte("content1")},
				"file2.txt": {Data: []byte("content2")},
			}, fstest.MapFS{}},
			dir:  ".",
			want: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "lower only",
			chain: []fs.FS{fstest.MapFS{}, fstest.MapFS{
				"file1.txt": {Data: []byte("content1")},
				"file2.txt": {Data: []byte("content2")},
			}},
			dir:  ".",
			want: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "layers merge and names are unique",
			chain: []fs.FS{
				fstest.MapFS{"file1.txt": {Data: []byte("upper1")}, "file2.txt": {Data: []byte("upper2")}},
				fstest.MapFS{"file2.txt": {Data: []byte("lower2")}, "file3.txt": {Data: []byte("lower3")}},
			},
			dir:  ".",
			want: []string{"file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name: "three layers merge",
			chain: []fs.FS{
				fstest.MapFS{"a.txt": {Data: []byte("a")}, "b.txt": {Data: []byte("b")}},
				fstest.MapFS{"b.txt": {Data: []byte("b2")}, "c.txt": {Data: []byte("c")}},
				fstest.MapFS{"c.txt": {Data: []byte("c2")}, "d.txt": {Data: []byte("d")}},
			},
			dir:  ".",
			want: []string{"a.txt", "b.txt", "c.txt", "d.txt"},
		},
		{
			name:  "results are sorted by name",
			chain: []fs.FS{fstest.MapFS{"zzz.txt": {Data: []byte("z")}, "aaa.txt": {Data: []byte("a")}}, fstest.MapFS{"mmm.txt": {Data: []byte("m")}}},
			dir:   ".",
			want:  []string{"aaa.txt", "mmm.txt", "zzz.txt"},
		},
		{
			name:  "a nil layer is skipped",
			chain: []fs.FS{fstest.MapFS{"file1.txt": {Data: []byte("content1")}}, nil},
			dir:   ".",
			want:  []string{"file1.txt"},
		},
		{
			name:  "one layer succeeding is enough",
			chain: []fs.FS{fstest.MapFS{}, fstest.MapFS{"file.txt": {Data: []byte("content")}}},
			dir:   ".",
			want:  []string{"file.txt"},
		},
		{
			name:  "both layers empty",
			chain: []fs.FS{fstest.MapFS{}, fstest.MapFS{}},
			dir:   ".",
			want:  []string{},
		},
		{
			name:  "every layer nil",
			chain: []fs.FS{nil, nil, nil},
			dir:   ".",
			want:  []string{},
		},
		{
			name:  "a directory no layer has is an error",
			chain: []fs.FS{fstest.MapFS{"file.txt": {Data: []byte("content")}}, fstest.MapFS{}},
			dir:   "nonexistent",
			err:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			overlay := vuego.NewOverlayFS(test.chain[0], test.chain[1:]...)

			entries, err := overlay.ReadDir(test.dir)
			if test.err {
				assert.Error(t, err)
				assert.Nil(t, entries)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.want, entryNames(entries))
		})
	}
}

func TestOverlayFS_Glob(t *testing.T) {
	tests := []struct {
		name    string
		chain   []fs.FS
		pattern string
		want    []string
	}{
		{
			name:    "upper only",
			chain:   []fs.FS{fstest.MapFS{"file1.txt": {Data: []byte("1")}, "file2.txt": {Data: []byte("2")}}, fstest.MapFS{}},
			pattern: "*.txt",
			want:    []string{"file1.txt", "file2.txt"},
		},
		{
			name:    "lower only",
			chain:   []fs.FS{fstest.MapFS{}, fstest.MapFS{"file1.txt": {Data: []byte("1")}, "file2.txt": {Data: []byte("2")}}},
			pattern: "*.txt",
			want:    []string{"file1.txt", "file2.txt"},
		},
		{
			name: "layers merge and matches are unique",
			chain: []fs.FS{
				fstest.MapFS{"file1.txt": {Data: []byte("upper1")}, "file2.txt": {Data: []byte("upper2")}},
				fstest.MapFS{"file2.txt": {Data: []byte("lower2")}, "file3.txt": {Data: []byte("lower3")}},
			},
			pattern: "*.txt",
			want:    []string{"file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name: "three layers merge",
			chain: []fs.FS{
				fstest.MapFS{"file1.txt": {Data: []byte("1")}, "file2.txt": {Data: []byte("2")}},
				fstest.MapFS{"file2.txt": {Data: []byte("2b")}, "file3.txt": {Data: []byte("3")}},
				fstest.MapFS{"file3.txt": {Data: []byte("3b")}, "file4.txt": {Data: []byte("4")}},
			},
			pattern: "*.txt",
			want:    []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"},
		},
		{
			name:    "results are sorted",
			chain:   []fs.FS{fstest.MapFS{"zzz.txt": {Data: []byte("z")}, "aaa.txt": {Data: []byte("a")}}, fstest.MapFS{"mmm.txt": {Data: []byte("m")}}},
			pattern: "*.txt",
			want:    []string{"aaa.txt", "mmm.txt", "zzz.txt"},
		},
		{
			name: "a prefix pattern does not descend",
			chain: []fs.FS{
				fstest.MapFS{
					"test1.txt":    {Data: []byte("1")},
					"test2.txt":    {Data: []byte("2")},
					"other.md":     {Data: []byte("o")},
					"sub/test.txt": {Data: []byte("sub")},
				},
				fstest.MapFS{"test3.txt": {Data: []byte("3")}},
			},
			pattern: "test*.txt",
			want:    []string{"test1.txt", "test2.txt", "test3.txt"},
		},
		{
			name:    "a nil layer is skipped",
			chain:   []fs.FS{fstest.MapFS{"file1.txt": {Data: []byte("1")}, "file2.txt": {Data: []byte("2")}}, nil},
			pattern: "*.txt",
			want:    []string{"file1.txt", "file2.txt"},
		},
		{
			name:    "no matches",
			chain:   []fs.FS{fstest.MapFS{"file.txt": {Data: []byte("content")}}, fstest.MapFS{}},
			pattern: "*.md",
			want:    []string{},
		},
		{
			name:    "every layer nil",
			chain:   []fs.FS{nil, nil, nil},
			pattern: "*.txt",
			want:    []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			overlay := vuego.NewOverlayFS(test.chain[0], test.chain[1:]...)

			matches, err := overlay.Glob(test.pattern)
			assert.NoError(t, err)
			assert.Equal(t, test.want, matches)
		})
	}
}
