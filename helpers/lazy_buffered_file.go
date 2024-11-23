package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"redun-pendancy/utils"
	"regexp"
	"strings"
)

// LazyBufferedFile is a file abstraction with lazy loading and in-memory buffering.
//
// It allows on-demand loading, modification, and optional writing back to disk.
//
// NOTE: This struct is NOT thread-safe; client code must ensure proper synchronization.
type LazyBufferedFile struct {
	FilePath string
	lines    []string
	loaded   bool
}

func NewLazyBufferedFile(filePath string) (*LazyBufferedFile, error) {
	_, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to access file at path %q: %v", filePath, err)
	}
	return &LazyBufferedFile{
		FilePath: filepath.Clean(filePath),
	}, nil
}

func (file *LazyBufferedFile) IsLoaded() bool {
	return file.loaded
}

func (file *LazyBufferedFile) GetLineCount() int {
	return len(file.lines)
}

// Retrieves the entire file content as a single string.
//
// NOTE: This operation constructs the content by joining all lines,
// which can be computationally expensive.
//
// Use this method sparingly.
func (file *LazyBufferedFile) GetContent() (string, error) {
	err := file.LoadOrSkip()
	if err != nil {
		return "", err
	}
	return strings.Join(file.lines, "\n"), nil
}

func (file *LazyBufferedFile) SetContent(content string) {
	lines := strings.Split(content, "\n")
	file.SetLines(lines)
}

func (file *LazyBufferedFile) GetLine(index int) (string, error) {
	err := file.LoadOrSkip()
	if err != nil {
		return "", err
	}
	return file.lines[index], nil
}

func (file *LazyBufferedFile) GetLines() ([]string, error) {
	err := file.LoadOrSkip()
	if err != nil {
		return nil, err
	}
	return file.lines, nil
}

func (file *LazyBufferedFile) LoadOrSkip() error {
	if file.loaded {
		return nil
	}
	err := file.Reload()
	return err
}

func (file *LazyBufferedFile) Reload() error {
	data, err := os.ReadFile(file.FilePath)
	if err != nil {
		return err
	}
	file.SetContent(string(data))
	return nil
}

func (file *LazyBufferedFile) SetLines(lines []string) {
	file.lines = lines
	file.loaded = true
}

func (file *LazyBufferedFile) SetLine(index int, lineContent string) error {
	err := file.LoadOrSkip()
	if err != nil {
		return err
	}

	file.lines[index] = lineContent
	return nil
}

func (file *LazyBufferedFile) InsertLine(index int, content string) error {
	err := file.LoadOrSkip()
	if err != nil {
		return err
	}

	file.lines = utils.InsertAt(file.lines, index, content)
	return nil
}

func (file *LazyBufferedFile) RemoveLine(index int) error {
	err := file.LoadOrSkip()
	if err != nil {
		return err
	}

	file.lines = utils.RemoveAt(file.lines, index)
	return nil
}

func (file *LazyBufferedFile) FindLineIndex(regex *regexp.Regexp, startIndex int) (int, error) {
	err := file.LoadOrSkip()
	if err != nil {
		return -1, err
	}

	index := utils.IndexOf(file.lines, startIndex, func(line string) bool {
		return regex.MatchString(line)
	})
	return index, nil
}

func (file *LazyBufferedFile) FindLastLineIndex(regex *regexp.Regexp, startIndex int) (int, error) {
	err := file.LoadOrSkip()
	if err != nil {
		return -1, err
	}

	index := utils.LastIndexOf(file.lines, startIndex, func(line string) bool {
		return regex.MatchString(line)
	})
	return index, nil
}

func (file *LazyBufferedFile) Commit() error {
	if !file.loaded {
		return fmt.Errorf("no content loaded to commit")
	}
	content := strings.Join(file.lines, "\n")
	return os.WriteFile(file.FilePath, []byte(content), 0644)
}

func (file *LazyBufferedFile) Unload() {
	file.loaded = false
	file.lines = nil
}
