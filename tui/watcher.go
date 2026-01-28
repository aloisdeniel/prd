package tui

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

type fileChangedMsg struct{}

func watchFiles(basePath string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil
		}
		// Don't close — this goroutine lives for the program lifetime.

		// Add basePath and all immediate subdirectories.
		addDirs(watcher, basePath)

		// Debounce: wait for quiet period before emitting.
		debounce := time.NewTimer(time.Hour)
		debounce.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					// If a new directory appeared, watch it too.
					if event.Op&fsnotify.Create != 0 {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							watcher.Add(event.Name)
						}
					}
					debounce.Reset(200 * time.Millisecond)
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
			case <-debounce.C:
				return fileChangedMsg{}
			}
		}
	}
}

func addDirs(watcher *fsnotify.Watcher, root string) {
	watcher.Add(root)
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			watcher.Add(filepath.Join(root, e.Name()))
		}
	}
}
