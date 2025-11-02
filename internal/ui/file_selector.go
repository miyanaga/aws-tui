package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rivo/tview"
)

type FileSelector struct {
	*tview.Flex
	directory      string
	table          *Table
	selectCallback func(string)
	files          []string
}

func NewFileSelector(directory string, selectCallback func(string)) *FileSelector {
	table := NewTable([]string{"FILE NAME", "SIZE"}, 1, 0)

	fs := &FileSelector{
		Flex:           tview.NewFlex(),
		directory:      directory,
		table:          table,
		selectCallback: selectCallback,
	}

	fs.Flex.SetDirection(tview.FlexRow)
	fs.Flex.SetBorder(true)
	fs.Flex.SetTitle(" Select File from " + directory + " ")
	fs.Flex.AddItem(table, 0, 1, true)

	fs.table.SetSelectedFunc(fs.onSelect)
	fs.loadFiles()

	return fs
}

func (fs *FileSelector) loadFiles() {
	entries, err := os.ReadDir(fs.directory)
	if err != nil {
		return
	}

	var data [][]string
	fs.files = make([]string, 0)

	// Collect files (skip directories and hidden files)
	for _, entry := range entries {
		if entry.IsDir() || entry.Name()[0] == '.' {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fs.files = append(fs.files, entry.Name())
		data = append(data, []string{
			entry.Name(),
			formatSize(info.Size()),
		})
	}

	// Sort by name
	sort.Slice(data, func(i, j int) bool {
		return data[i][0] < data[j][0]
	})
	sort.Strings(fs.files)

	fs.table.SetData(data)
	if len(data) > 0 {
		fs.table.Select(1, 0)
	}
}

func (fs *FileSelector) onSelect(row, col int) {
	if row <= 0 || row > len(fs.files) {
		return
	}

	filename := fs.files[row-1]
	fullPath := filepath.Join(fs.directory, filename)

	if fs.selectCallback != nil {
		fs.selectCallback(fullPath)
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}
