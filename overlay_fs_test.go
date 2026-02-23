package vuego_test

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/vuego"
)

var NewOverlayFS = vuego.NewOverlayFS

func TestOverlayFSOpen_UpperOnly(t *testing.T) {
	upper := fstest.MapFS{
		"file.txt": {Data: []byte("upper content")},
	}
	lower := os.DirFS("none")

	o := NewOverlayFS(upper, lower)
	f, err := o.Open("file.txt")
	assert.NoError(t, err)
	defer f.Close()

	data := make([]byte, 13)
	n, err := f.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, "upper content", string(data))
}

func TestOverlayFSOpen_LowerFallback(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{
		"file.txt": {Data: []byte("lower content")},
	}

	o := NewOverlayFS(upper, lower)
	f, err := o.Open("file.txt")
	assert.NoError(t, err)
	defer f.Close()

	data := make([]byte, 13)
	n, err := f.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, "lower content", string(data))
}

func TestOverlayFSOpen_UpperOverrides(t *testing.T) {
	upper := fstest.MapFS{
		"file.txt": {Data: []byte("upper content")},
	}
	lower := fstest.MapFS{
		"file.txt": {Data: []byte("lower content")},
	}

	o := NewOverlayFS(upper, lower)
	f, err := o.Open("file.txt")
	assert.NoError(t, err)
	defer f.Close()

	data := make([]byte, 13)
	n, err := f.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, "upper content", string(data))
}

func TestOverlayFSOpen_NotFound(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	_, err := o.Open("file.txt")
	assert.Error(t, err)
}

func TestOverlayFSReadDir_UpperOnly(t *testing.T) {
	upper := fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
		"file2.txt": {Data: []byte("content2")},
	}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)

	names := extractNames(entries)
	assert.Len(t, names, 2)
	assert.Contains(t, names, "file1.txt")
	assert.Contains(t, names, "file2.txt")
}

func TestOverlayFSReadDir_LowerOnly(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
		"file2.txt": {Data: []byte("content2")},
	}

	o := NewOverlayFS(upper, lower)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)

	names := extractNames(entries)
	assert.Len(t, names, 2)
	assert.Contains(t, names, "file1.txt")
	assert.Contains(t, names, "file2.txt")
}

func TestOverlayFSReadDir_Merged(t *testing.T) {
	upper := fstest.MapFS{
		"file1.txt": {Data: []byte("upper1")},
		"file2.txt": {Data: []byte("upper2")},
	}
	lower := fstest.MapFS{
		"file2.txt": {Data: []byte("lower2")},
		"file3.txt": {Data: []byte("lower3")},
	}

	o := NewOverlayFS(upper, lower)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)

	names := extractNames(entries)
	assert.Len(t, names, 3)
	assert.Contains(t, names, "file1.txt")
	assert.Contains(t, names, "file2.txt")
	assert.Contains(t, names, "file3.txt")
}

func TestOverlayFSReadDir_BothEmpty(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestOverlayFSReadDir_InvalidPath(t *testing.T) {
	upper := fstest.MapFS{
		"file.txt": {Data: []byte("content")},
	}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	_, err := o.ReadDir("nonexistent")
	assert.Error(t, err)
}

func TestOverlayFSGlob_UpperOnly(t *testing.T) {
	upper := fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
		"file2.txt": {Data: []byte("content2")},
	}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 2)
	assert.Contains(t, matches, "file1.txt")
	assert.Contains(t, matches, "file2.txt")
}

func TestOverlayFSGlob_LowerOnly(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
		"file2.txt": {Data: []byte("content2")},
	}

	o := NewOverlayFS(upper, lower)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 2)
	assert.Contains(t, matches, "file1.txt")
	assert.Contains(t, matches, "file2.txt")
}

func TestOverlayFSGlob_Merged(t *testing.T) {
	upper := fstest.MapFS{
		"file1.txt": {Data: []byte("upper1")},
		"file2.txt": {Data: []byte("upper2")},
	}
	lower := fstest.MapFS{
		"file2.txt": {Data: []byte("lower2")},
		"file3.txt": {Data: []byte("lower3")},
	}

	o := NewOverlayFS(upper, lower)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 3)
	assert.Contains(t, matches, "file1.txt")
	assert.Contains(t, matches, "file2.txt")
	assert.Contains(t, matches, "file3.txt")
}

func TestOverlayFSGlob_NoMatches(t *testing.T) {
	upper := fstest.MapFS{
		"file.txt": {Data: []byte("content")},
	}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	matches, err := o.Glob("*.md")
	assert.NoError(t, err)

	assert.Empty(t, matches)
}

func TestOverlayFSGlob_Sorted(t *testing.T) {
	upper := fstest.MapFS{
		"zzz.txt": {Data: []byte("z")},
		"aaa.txt": {Data: []byte("a")},
	}
	lower := fstest.MapFS{
		"mmm.txt": {Data: []byte("m")},
	}

	o := NewOverlayFS(upper, lower)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 3)
	// Check if sorted
	assert.IsIncreasing(t, matches)
}

func TestOverlayFSReadDir_Sorted(t *testing.T) {
	upper := fstest.MapFS{
		"zzz.txt": {Data: []byte("z")},
		"aaa.txt": {Data: []byte("a")},
	}
	lower := fstest.MapFS{
		"mmm.txt": {Data: []byte("m")},
	}

	o := NewOverlayFS(upper, lower)
	dirEntries, err := o.ReadDir(".")
	assert.NoError(t, err)

	assert.Len(t, dirEntries, 3)
	names := extractNames(dirEntries)
	assert.IsIncreasing(t, names)
}

func TestOverlayFSOpenWithNestedPath(t *testing.T) {
	upper := fstest.MapFS{
		"dir/file.txt": {Data: []byte("upper")},
	}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	f, err := o.Open("dir/file.txt")
	assert.NoError(t, err)
	defer f.Close()

	data := make([]byte, 5)
	n, err := f.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "upper", string(data[:n]))
}

func TestOverlayFSOpen_WithNilFS(t *testing.T) {
	upper := fstest.MapFS{
		"file.txt": {Data: []byte("upper content")},
	}

	// Create overlay with nil in chain
	o := NewOverlayFS(upper, nil)
	f, err := o.Open("file.txt")
	assert.NoError(t, err)
	defer f.Close()

	data := make([]byte, 13)
	n, err := f.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, "upper content", string(data))
}

func TestOverlayFSOpen_MultipleChains(t *testing.T) {
	fs1 := fstest.MapFS{}
	fs2 := fstest.MapFS{}
	fs3 := fstest.MapFS{
		"file.txt": {Data: []byte("found in fs3")},
	}

	o := NewOverlayFS(fs1, fs2, fs3)
	f, err := o.Open("file.txt")
	assert.NoError(t, err)
	defer f.Close()

	data := make([]byte, 12)
	n, err := f.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, "found in fs3", string(data))
}

func TestOverlayFSReadDir_WithNilFS(t *testing.T) {
	upper := fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
	}

	o := NewOverlayFS(upper, nil)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)

	names := extractNames(entries)
	assert.Len(t, names, 1)
	assert.Contains(t, names, "file1.txt")
}

func TestOverlayFSReadDir_MultipleChainsMerge(t *testing.T) {
	fs1 := fstest.MapFS{
		"a.txt": {Data: []byte("a")},
		"b.txt": {Data: []byte("b")},
	}
	fs2 := fstest.MapFS{
		"b.txt": {Data: []byte("b2")},
		"c.txt": {Data: []byte("c")},
	}
	fs3 := fstest.MapFS{
		"c.txt": {Data: []byte("c2")},
		"d.txt": {Data: []byte("d")},
	}

	o := NewOverlayFS(fs1, fs2, fs3)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)

	names := extractNames(entries)
	assert.Len(t, names, 4)
	assert.Contains(t, names, "a.txt")
	assert.Contains(t, names, "b.txt")
	assert.Contains(t, names, "c.txt")
	assert.Contains(t, names, "d.txt")
}

func TestOverlayFSReadDir_AllFSHaveError(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{}

	o := NewOverlayFS(upper, lower)
	entries, err := o.ReadDir("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, entries)
}

func TestOverlayFSReadDir_OneFSSucceedsOneErrors(t *testing.T) {
	upper := fstest.MapFS{}
	lower := fstest.MapFS{
		"file.txt": {Data: []byte("content")},
	}

	o := NewOverlayFS(upper, lower)
	entries, err := o.ReadDir(".")
	assert.NoError(t, err)

	names := extractNames(entries)
	assert.Len(t, names, 1)
	assert.Contains(t, names, "file.txt")
}

func TestOverlayFSGlob_WithNilFS(t *testing.T) {
	upper := fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
		"file2.txt": {Data: []byte("content2")},
	}

	o := NewOverlayFS(upper, nil)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 2)
	assert.Contains(t, matches, "file1.txt")
	assert.Contains(t, matches, "file2.txt")
}

func TestOverlayFSGlob_MultipleChainsMerge(t *testing.T) {
	fs1 := fstest.MapFS{
		"file1.txt": {Data: []byte("1")},
		"file2.txt": {Data: []byte("2")},
	}
	fs2 := fstest.MapFS{
		"file2.txt": {Data: []byte("2b")},
		"file3.txt": {Data: []byte("3")},
	}
	fs3 := fstest.MapFS{
		"file3.txt": {Data: []byte("3b")},
		"file4.txt": {Data: []byte("4")},
	}

	o := NewOverlayFS(fs1, fs2, fs3)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 4)
	assert.Contains(t, matches, "file1.txt")
	assert.Contains(t, matches, "file2.txt")
	assert.Contains(t, matches, "file3.txt")
	assert.Contains(t, matches, "file4.txt")
}

func TestOverlayFSGlob_ComplexPattern(t *testing.T) {
	upper := fstest.MapFS{
		"test1.txt":    {Data: []byte("1")},
		"test2.txt":    {Data: []byte("2")},
		"other.md":     {Data: []byte("o")},
		"sub/test.txt": {Data: []byte("sub")},
	}
	lower := fstest.MapFS{
		"test3.txt": {Data: []byte("3")},
	}

	o := NewOverlayFS(upper, lower)
	matches, err := o.Glob("test*.txt")
	assert.NoError(t, err)

	assert.Len(t, matches, 3)
	assert.Contains(t, matches, "test1.txt")
	assert.Contains(t, matches, "test2.txt")
	assert.Contains(t, matches, "test3.txt")
}

func TestOverlayFSOpen_AllFSSNil(t *testing.T) {
	o := NewOverlayFS(nil, nil, nil)
	_, err := o.Open("file.txt")
	assert.Error(t, err)
}

func TestOverlayFSReadDir_AllFSNil(t *testing.T) {
	o := NewOverlayFS(nil, nil, nil)
	entries, err := o.ReadDir(".")
	// Should return empty list, not error when all FSs are nil
	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestOverlayFSGlob_AllFSNil(t *testing.T) {
	o := NewOverlayFS(nil, nil, nil)
	matches, err := o.Glob("*.txt")
	assert.NoError(t, err)
	assert.Empty(t, matches)
}

// Helper functions
func extractNames(entries []fs.DirEntry) []string {
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
