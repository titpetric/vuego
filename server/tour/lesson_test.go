package tour_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/server/tour"
)

func TestParseTour(t *testing.T) {
	fs := fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Test Chapter

## First Lesson

Lesson content here.

@file: first.vuego

---

## Second Lesson

More content.

@file: second.vuego
`),
		},
		"basics/first.vuego": &fstest.MapFile{
			Data: []byte(`<div>Hello</div>`),
		},
		"basics/first.json": &fstest.MapFile{
			Data: []byte(`{"name": "test"}`),
		},
		"basics/second.vuego": &fstest.MapFile{
			Data: []byte(`<div>World</div>`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)
	require.Len(t, parsed.Chapters, 1)

	chapter := parsed.Chapters[0]
	require.Equal(t, "01-basics", chapter.Name)
	require.Equal(t, "Test Chapter", chapter.Title)
	require.Len(t, chapter.Lessons, 2)

	lesson1 := chapter.Lessons[0]
	require.Equal(t, "0/0", lesson1.ID)
	require.Equal(t, "First Lesson", lesson1.Title)
	require.Equal(t, "Lesson content here.", lesson1.Content)
	require.Equal(t, 0, lesson1.ChapterIdx)
	require.Equal(t, 0, lesson1.LessonIdx)
	require.Contains(t, lesson1.Files, "first.vuego")
	require.Contains(t, lesson1.Files, "first.json") // Implicitly loaded

	lesson2 := chapter.Lessons[1]
	require.Equal(t, "0/1", lesson2.ID)
	require.Equal(t, "Second Lesson", lesson2.Title)
	require.Equal(t, "More content.", lesson2.Content)
}

func TestParseTour_MultipleChapters(t *testing.T) {
	fs := fstest.MapFS{
		"01-intro.md": &fstest.MapFile{
			Data: []byte(`# Introduction

## Welcome

Welcome content.

@file: index.vuego
`),
		},
		"intro/index.vuego": &fstest.MapFile{
			Data: []byte(`<div>Welcome</div>`),
		},
		"02-advanced.md": &fstest.MapFile{
			Data: []byte(`# Advanced

## Deep Dive

Advanced content.

@file: index.vuego
`),
		},
		"advanced/index.vuego": &fstest.MapFile{
			Data: []byte(`<div>Advanced</div>`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)
	require.Len(t, parsed.Chapters, 2)
	require.Equal(t, "Introduction", parsed.Chapters[0].Title)
	require.Equal(t, "Advanced", parsed.Chapters[1].Title)
}

func TestTour_GetLesson(t *testing.T) {
	fs := fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Basics

## First Lesson

Content one.

@file: first.vuego

---

## Second Lesson

Content two.

@file: second.vuego
`),
		},
		"basics/first.vuego": &fstest.MapFile{
			Data: []byte(`<div>First</div>`),
		},
		"basics/second.vuego": &fstest.MapFile{
			Data: []byte(`<div>Second</div>`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)

	lesson := parsed.GetLesson("0/1")
	require.NotNil(t, lesson)
	require.Equal(t, "Second Lesson", lesson.Title)
	require.Equal(t, "Content two.", lesson.Content)
}

func TestTour_GetLesson_NotFound(t *testing.T) {
	fs := fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Basics

## First Lesson

Content.

@file: first.vuego
`),
		},
		"basics/first.vuego": &fstest.MapFile{
			Data: []byte(`<div>First</div>`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)

	lesson := parsed.GetLesson("99/99")
	require.Nil(t, lesson)
}

func TestTour_FirstLesson(t *testing.T) {
	fs := fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Basics

## First Lesson

First content.

@file: first.vuego
## Second Lesson

Second content.

@file: second.vuego
`),
		},
		"basics/first.vuego": &fstest.MapFile{
			Data: []byte(`<div>First</div>`),
		},
		"basics/second.vuego": &fstest.MapFile{
			Data: []byte(`<div>Second</div>`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)

	first := parsed.FirstLesson()
	require.NotNil(t, first)
	require.Equal(t, "First Lesson", first.Title)
	require.Equal(t, "0/0", first.ID)
}

func TestTour_FirstLesson_Empty(t *testing.T) {
	fs := fstest.MapFS{
		"01-empty.md": &fstest.MapFile{
			Data: []byte(`# Empty Chapter
`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)
	require.Nil(t, parsed.FirstLesson())
}

func TestTour_LessonCount(t *testing.T) {
	fs := fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Basics

## Lesson One

Content.

@file: one.vuego

---

## Lesson Two

Content.

@file: two.vuego

---

## Lesson Three

Content.

@file: three.vuego
`),
		},
		"basics/one.vuego":   &fstest.MapFile{Data: []byte(`<div>1</div>`)},
		"basics/two.vuego":   &fstest.MapFile{Data: []byte(`<div>2</div>`)},
		"basics/three.vuego": &fstest.MapFile{Data: []byte(`<div>3</div>`)},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)
	require.Equal(t, 3, parsed.LessonCount())
}

func TestTour_LessonNavigation(t *testing.T) {
	fs := fstest.MapFS{
		"01-basics.md": &fstest.MapFile{
			Data: []byte(`# Basics

## First

Content.

---

## Second

Content.

---

## Third

Content.
`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)

	first := parsed.GetLesson("0/0")
	require.NotNil(t, first)
	require.False(t, first.HasPrev)
	require.True(t, first.HasNext)
	require.Equal(t, "", first.PrevID)
	require.Equal(t, "0/1", first.NextID)

	second := parsed.GetLesson("0/1")
	require.NotNil(t, second)
	require.True(t, second.HasPrev)
	require.True(t, second.HasNext)
	require.Equal(t, "0/0", second.PrevID)
	require.Equal(t, "0/2", second.NextID)

	third := parsed.GetLesson("0/2")
	require.NotNil(t, third)
	require.True(t, third.HasPrev)
	require.False(t, third.HasNext)
	require.Equal(t, "0/1", third.PrevID)
	require.Equal(t, "", third.NextID)
}

func TestTour_LessonNavigation_AcrossChapters(t *testing.T) {
	fs := fstest.MapFS{
		"01-intro.md": &fstest.MapFile{
			Data: []byte(`# Intro

## Intro Lesson

Content.
`),
		},
		"02-advanced.md": &fstest.MapFile{
			Data: []byte(`# Advanced

## Advanced Lesson

Content.
`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)

	introLesson := parsed.GetLesson("0/0")
	require.NotNil(t, introLesson)
	require.True(t, introLesson.HasNext)
	require.Equal(t, "1/0", introLesson.NextID)

	advancedLesson := parsed.GetLesson("1/0")
	require.NotNil(t, advancedLesson)
	require.True(t, advancedLesson.HasPrev)
	require.Equal(t, "0/0", advancedLesson.PrevID)
}

func TestLesson_PrimaryTemplate(t *testing.T) {
	t.Run("returns index.vuego when present", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: index.vuego
@file: other.vuego
`),
			},
			"basics/index.vuego": &fstest.MapFile{
				Data: []byte(`<div>Hello</div>`),
			},
			"basics/other.vuego": &fstest.MapFile{
				Data: []byte(`<div>Other</div>`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)
		require.Len(t, parsed.Chapters, 1)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "index.vuego", lesson.PrimaryTemplate())
	})

	t.Run("returns first vuego file when no index.vuego", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: template.vuego
`),
			},
			"basics/template.vuego": &fstest.MapFile{
				Data: []byte(`<div>Hello</div>`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "template.vuego", lesson.PrimaryTemplate())
	})

	t.Run("returns empty string when no vuego files", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: data.json
`),
			},
			"basics/data.json": &fstest.MapFile{
				Data: []byte(`{"name": "test"}`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "", lesson.PrimaryTemplate())
	})
}

func TestLesson_DataFile(t *testing.T) {
	t.Run("returns index.yml when index.vuego and index.yml exist", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: index.vuego
@file: index.yml
`),
			},
			"basics/index.vuego": &fstest.MapFile{
				Data: []byte(`<div>Hello</div>`),
			},
			"basics/index.yml": &fstest.MapFile{
				Data: []byte(`name: test`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "index.yml", lesson.DataFile())
	})

	t.Run("returns index.json when index.vuego and index.json exist", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: index.vuego
@file: index.json
`),
			},
			"basics/index.vuego": &fstest.MapFile{
				Data: []byte(`<div>Hello</div>`),
			},
			"basics/index.json": &fstest.MapFile{
				Data: []byte(`{"name": "test"}`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "index.json", lesson.DataFile())
	})

	t.Run("returns empty string when no matching data file", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: index.vuego
@file: other.json
`),
			},
			"basics/index.vuego": &fstest.MapFile{
				Data: []byte(`<div>Hello</div>`),
			},
			"basics/other.json": &fstest.MapFile{
				Data: []byte(`{"name": "test"}`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "", lesson.DataFile())
	})

	t.Run("returns empty string when no primary template", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: index.json
`),
			},
			"basics/index.json": &fstest.MapFile{
				Data: []byte(`{"name": "test"}`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Equal(t, "", lesson.DataFile())
	})
}

func TestParseTour_LoadsLessonFiles(t *testing.T) {
	t.Run("lesson files are loaded from @file references", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: template.vuego
@file: data.json
`),
			},
			"basics/template.vuego": &fstest.MapFile{
				Data: []byte(`<div>Hello</div>`),
			},
			"basics/data.json": &fstest.MapFile{
				Data: []byte(`{"name": "test"}`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Len(t, lesson.Files, 2)
		require.Equal(t, "<div>Hello</div>", lesson.Files["template.vuego"])
		require.Equal(t, `{"name": "test"}`, lesson.Files["data.json"])
	})

	t.Run("multiple files per lesson are loaded", func(t *testing.T) {
		fs := fstest.MapFS{
			"01-basics.md": &fstest.MapFile{
				Data: []byte(`# Basics

## First Lesson

Content.

@file: index.vuego
@file: layout.vuego
@file: index.yml
`),
			},
			"basics/index.vuego": &fstest.MapFile{
				Data: []byte(`<div>Main</div>`),
			},
			"basics/layout.vuego": &fstest.MapFile{
				Data: []byte(`<html><slot /></html>`),
			},
			"basics/index.yml": &fstest.MapFile{
				Data: []byte(`name: test`),
			},
		}

		parsed, err := tour.ParseTour(fs)
		require.NoError(t, err)

		lesson := parsed.GetLesson("0/0")
		require.NotNil(t, lesson)
		require.Len(t, lesson.Files, 3)
		require.Equal(t, "<div>Main</div>", lesson.Files["index.vuego"])
		require.Equal(t, "<html><slot /></html>", lesson.Files["layout.vuego"])
		require.Equal(t, "name: test", lesson.Files["index.yml"])
	})
}

func TestParseTour_EmbeddedContent(t *testing.T) {
	// Parse the embedded content directly
	parsed, err := tour.ParseTour(tour.EmbeddedContentFS())
	require.NoError(t, err)

	// Validate all lessons have .vuego templates
	require.NoError(t, tour.ValidateTour(parsed))

	require.Len(t, parsed.Chapters, 4, "expected 4 chapters in embedded content")

	// Verify chapters and lesson counts based on ## headings in content files
	expectedChapters := []struct {
		title       string
		lessonCount int
	}{
		{"Variable Interpolation", 3}, // 01-interpolation.md has 3 ## headings
		{"Filters", 3},                // 02-filters.md has 3 ## headings
		{"Directives", 4},             // 03-directives.md has 4 ## headings
		{"Components", 5},             // 04-components.md has 5 ## headings
	}

	for i, expected := range expectedChapters {
		require.Equal(t, expected.title, parsed.Chapters[i].Title, "chapter %d title mismatch", i)
		require.Len(t, parsed.Chapters[i].Lessons, expected.lessonCount, "chapter %d lesson count mismatch", i)
	}
}

func TestParseTour_DelimiterParsing(t *testing.T) {
	fs := fstest.MapFS{
		"01-test.md": &fstest.MapFile{
			Data: []byte(`# Chapter Title

## First Lesson

Content here.

@file: first.vuego

---

## Second Lesson

More content.

@file: second.vuego
`),
		},
		"test/first.vuego": &fstest.MapFile{
			Data: []byte(`<div>First</div>`),
		},
		"test/second.vuego": &fstest.MapFile{
			Data: []byte(`<div>Second</div>`),
		},
	}

	parsed, err := tour.ParseTour(fs)
	require.NoError(t, err)
	require.Len(t, parsed.Chapters, 1)

	chapter := parsed.Chapters[0]
	require.Equal(t, "Chapter Title", chapter.Title)
	require.Len(t, chapter.Lessons, 2)

	require.Equal(t, "First Lesson", chapter.Lessons[0].Title)
	require.Equal(t, "Second Lesson", chapter.Lessons[1].Title)
}

func TestLesson_PrimaryTemplate_Direct(t *testing.T) {
	t.Run("index.vuego returns index.vuego", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"index.vuego": "<div>Hello</div>",
				"other.vuego": "<div>Other</div>",
			},
		}
		require.Equal(t, "index.vuego", lesson.PrimaryTemplate())
	})

	t.Run("only other.vuego returns other.vuego", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"other.vuego": "<div>Other</div>",
			},
		}
		require.Equal(t, "other.vuego", lesson.PrimaryTemplate())
	})

	t.Run("no vuego files returns empty string", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"data.json": `{"key": "value"}`,
			},
		}
		require.Equal(t, "", lesson.PrimaryTemplate())
	})
}

func TestLesson_DataFile_Direct(t *testing.T) {
	t.Run("index.vuego + index.yml returns index.yml", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"index.vuego": "<div>Hello</div>",
				"index.yml":   "key: value",
			},
		}
		require.Equal(t, "index.yml", lesson.DataFile())
	})

	t.Run("index.vuego + index.json returns index.json", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"index.vuego": "<div>Hello</div>",
				"index.json":  `{"key": "value"}`,
			},
		}
		require.Equal(t, "index.json", lesson.DataFile())
	})

	t.Run("index.vuego + no data file returns empty string", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"index.vuego": "<div>Hello</div>",
			},
		}
		require.Equal(t, "", lesson.DataFile())
	})

	t.Run("no primary template returns empty string", func(t *testing.T) {
		lesson := &tour.Lesson{
			Files: map[string]string{
				"index.json": `{"key": "value"}`,
			},
		}
		require.Equal(t, "", lesson.DataFile())
	})
}
