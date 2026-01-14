package tour

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Lesson represents a single lesson within a chapter.
type Lesson struct {
	ID             string            // Unique identifier (chapter/lesson index)
	Title          string            // Lesson title from ## heading
	Content        string            // Markdown content for the lesson
	Files          map[string]string // Template and data files for this lesson
	FileRefs       []string          // File paths from @file: references
	Chapter        string            // Parent chapter name
	ChapterTitle   string            // Parent chapter title (friendly name)
	ChapterSlug    string            // URL-friendly chapter name (e.g., "interpolation")
	ChapterIdx     int               // Chapter index (0-based)
	LessonIdx      int               // Lesson index within chapter (0-based)
	HasPrev        bool              // Whether there's a previous lesson
	HasNext        bool              // Whether there's a next lesson
	PrevID         string            // Previous lesson ID
	NextID         string            // Next lesson ID
	PrevSlug       string            // Previous lesson chapter slug
	PrevLessonIdx  int               // Previous lesson index
	NextSlug       string            // Next lesson chapter slug
	NextLessonIdx  int               // Next lesson index
	TotalInChapter int               // Total lessons in this chapter
}

// Chapter represents a chapter containing multiple lessons.
type Chapter struct {
	Name    string    // Chapter name (from filename)
	Title   string    // Chapter title (from # heading or filename)
	Slug    string    // URL-friendly name (e.g., "interpolation")
	Lessons []*Lesson // Lessons in this chapter
	Index   int       // Chapter index (0-based)
}

// Tour holds all chapters and lessons.
type Tour struct {
	Chapters []*Chapter
	lessons  []*Lesson // Flat list for navigation
}

// ParseTour parses a tour from the given filesystem.
// It expects markdown files in the root and lesson files in subdirectories.
func ParseTour(contentFS fs.FS) (*Tour, error) {
	tour := &Tour{}

	// Find all markdown files (chapters)
	entries, err := fs.ReadDir(contentFS, ".")
	if err != nil {
		return nil, fmt.Errorf("reading tour directory: %w", err)
	}

	var mdFiles []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip README.md, DONE.md and other non-chapter files
		if !entry.IsDir() && strings.HasSuffix(name, ".md") && name != "README.md" && name != "DONE.md" {
			mdFiles = append(mdFiles, name)
		}
	}

	sort.Strings(mdFiles)

	for chapterIdx, mdFile := range mdFiles {
		chapter, err := parseChapter(contentFS, mdFile, chapterIdx)
		if err != nil {
			return nil, fmt.Errorf("parsing chapter %s: %w", mdFile, err)
		}
		tour.Chapters = append(tour.Chapters, chapter)
		tour.lessons = append(tour.lessons, chapter.Lessons...)
	}

	// Set navigation links
	for i, lesson := range tour.lessons {
		if i > 0 {
			prev := tour.lessons[i-1]
			lesson.HasPrev = true
			lesson.PrevID = prev.ID
			lesson.PrevSlug = prev.ChapterSlug
			lesson.PrevLessonIdx = prev.LessonIdx
		}
		if i < len(tour.lessons)-1 {
			next := tour.lessons[i+1]
			lesson.HasNext = true
			lesson.NextID = next.ID
			lesson.NextSlug = next.ChapterSlug
			lesson.NextLessonIdx = next.LessonIdx
		}
	}

	return tour, nil
}

func parseChapter(contentFS fs.FS, mdFile string, chapterIdx int) (*Chapter, error) {
	content, err := fs.ReadFile(contentFS, mdFile)
	if err != nil {
		return nil, err
	}

	chapterName := strings.TrimSuffix(mdFile, ".md")
	chapter := &Chapter{
		Name:  chapterName,
		Title: chapterName,
		Slug:  chapterDir(chapterName),
		Index: chapterIdx,
	}

	// Parse markdown into lessons (split by --- delimiter)
	lessons, chapterTitle := parseMarkdownLessons(string(content), chapterName, chapterIdx)
	if chapterTitle != "" {
		chapter.Title = chapterTitle
	}

	// Load files for each lesson from @file references
	for _, lesson := range lessons {
		lesson.Files = loadLessonFilesFromRefs(contentFS, chapterName, lesson.FileRefs)
		lesson.TotalInChapter = len(lessons)
		lesson.ChapterTitle = chapter.Title
		lesson.ChapterSlug = chapterDir(chapterName)
	}

	chapter.Lessons = lessons
	return chapter, nil
}

// ValidateTour checks that all lessons in the tour have valid content.
func ValidateTour(t *Tour) error {
	for _, lesson := range t.lessons {
		if err := ValidateLesson(lesson); err != nil {
			return err
		}
	}
	return nil
}

func parseMarkdownLessons(content, chapterName string, chapterIdx int) ([]*Lesson, string) {
	var lessons []*Lesson
	var chapterTitle string

	// Split by lesson delimiter: \n\n---\n\n
	sections := strings.Split(content, "\n\n---\n\n")

	for lessonIdx, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		// First section may contain chapter title (# heading)
		lines := strings.SplitN(section, "\n", 2)
		title := ""
		lessonContent := section

		if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
			// This is the chapter title, extract it
			if chapterTitle == "" {
				chapterTitle = strings.TrimPrefix(lines[0], "# ")
			}
			if len(lines) > 1 {
				lessonContent = strings.TrimSpace(lines[1])
			} else {
				continue // Only chapter title, no lesson content
			}
		}

		// Extract lesson title from ## heading if present
		contentLines := strings.SplitN(lessonContent, "\n", 2)
		if len(contentLines) > 0 && strings.HasPrefix(contentLines[0], "## ") {
			title = strings.TrimPrefix(contentLines[0], "## ")
			if len(contentLines) > 1 {
				lessonContent = strings.TrimSpace(contentLines[1])
			} else {
				lessonContent = ""
			}
		}

		if title == "" {
			title = fmt.Sprintf("Lesson %d", lessonIdx+1)
		}

		// Extract @file: references from content
		fileRefs, cleanContent := extractFileRefs(lessonContent)

		lesson := &Lesson{
			ID:         fmt.Sprintf("%d/%d", chapterIdx, len(lessons)),
			Title:      title,
			Content:    cleanContent,
			FileRefs:   fileRefs,
			Chapter:    chapterName,
			ChapterIdx: chapterIdx,
			LessonIdx:  len(lessons),
			Files:      make(map[string]string),
		}
		lessons = append(lessons, lesson)
	}

	return lessons, chapterTitle
}

// extractFileRefs extracts @file: references from lesson content.
// Returns the list of file paths and the content with @file: lines removed.
func extractFileRefs(content string) ([]string, string) {
	var refs []string
	var cleanLines []string

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@file:") {
			filePath := strings.TrimSpace(strings.TrimPrefix(trimmed, "@file:"))
			if filePath != "" {
				refs = append(refs, filePath)
			}
		} else {
			cleanLines = append(cleanLines, line)
		}
	}

	return refs, strings.TrimSpace(strings.Join(cleanLines, "\n"))
}

// chapterDir returns the directory name for a chapter (strips numeric prefix).
// e.g., "01-interpolation" -> "interpolation"
func chapterDir(chapterName string) string {
	parts := strings.SplitN(chapterName, "-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return chapterName
}

// loadLessonFilesFromRefs loads files from @file: references.
// It also implicitly loads data files (.yml, .yaml, .json) based on the primary template name.
func loadLessonFilesFromRefs(contentFS fs.FS, chapterName string, refs []string) map[string]string {
	files := make(map[string]string)
	dir := chapterDir(chapterName)

	for _, ref := range refs {
		// Build full path relative to chapter directory (without numeric prefix)
		filePath := path.Join(dir, ref)
		content, err := fs.ReadFile(contentFS, filePath)
		if err != nil {
			continue
		}
		// Use just the filename as the key
		fileName := path.Base(ref)
		files[fileName] = string(content)

		// Implicitly load data file for .vuego templates
		if strings.HasSuffix(fileName, ".vuego") {
			basePath := strings.TrimSuffix(filePath, ".vuego")
			baseName := strings.TrimSuffix(fileName, ".vuego")
			for _, ext := range []string{".yaml", ".yml", ".json"} {
				dataPath := basePath + ext
				dataContent, err := fs.ReadFile(contentFS, dataPath)
				if err == nil {
					files[baseName+ext] = string(dataContent)
					break
				}
			}
		}
	}

	return files
}

// ValidateLesson checks that a lesson has at least one .vuego file.
func ValidateLesson(lesson *Lesson) error {
	for _, ref := range lesson.FileRefs {
		if strings.HasSuffix(ref, ".vuego") {
			return nil
		}
	}
	return fmt.Errorf("lesson %q in chapter %q has no .vuego template file", lesson.Title, lesson.Chapter)
}

// PrimaryTemplate returns the main .vuego template filename for the lesson.
// It looks for index.vuego first, then any .vuego file.
func (l *Lesson) PrimaryTemplate() string {
	if _, ok := l.Files["index.vuego"]; ok {
		return "index.vuego"
	}
	for name := range l.Files {
		if strings.HasSuffix(name, ".vuego") {
			return name
		}
	}
	return ""
}

// DataFile returns the data filename (.yaml or .json) that matches the primary template.
func (l *Lesson) DataFile() string {
	primary := l.PrimaryTemplate()
	if primary == "" {
		return ""
	}
	base := strings.TrimSuffix(primary, ".vuego")
	for _, ext := range []string{".yaml", ".yml", ".json"} {
		if _, ok := l.Files[base+ext]; ok {
			return base + ext
		}
	}
	return ""
}

// GetLesson returns a lesson by its ID.
func (t *Tour) GetLesson(id string) *Lesson {
	for _, lesson := range t.lessons {
		if lesson.ID == id {
			return lesson
		}
	}
	return nil
}

// GetLessonByName returns a lesson by chapter name and lesson index.
// The chapter name is matched against the suffix of the chapter name (e.g., "interpolation" matches "01-interpolation").
func (t *Tour) GetLessonByName(chapterName, lessonIdx string) *Lesson {
	for _, chapter := range t.Chapters {
		// Match friendly name to prefixed chapter name
		if chapterDir(chapter.Name) == chapterName {
			idx, err := strconv.Atoi(lessonIdx)
			if err != nil || idx < 0 || idx >= len(chapter.Lessons) {
				return nil
			}
			return chapter.Lessons[idx]
		}
	}
	return nil
}

// FirstLesson returns the first lesson in the tour.
func (t *Tour) FirstLesson() *Lesson {
	if len(t.lessons) > 0 {
		return t.lessons[0]
	}
	return nil
}

// LessonCount returns the total number of lessons.
func (t *Tour) LessonCount() int {
	return len(t.lessons)
}
