package runner

import (
	"context"
	"os"
	"time"
	"fmt"

	"github.com/joshuarli/filewatcher/files"
	"github.com/fsnotify/fsnotify"
)

// WatchOptions passed to watch
type WatchOptions struct {
	IdleTimeout time.Duration
	Runner      *Runner
}

// Watch for events from the watcher and handle them with the runner
func Watch(watcher *fsnotify.Watcher, opts WatchOptions) error {
	runner := opts.Runner
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runner.start(ctx)

	for {
		select {
		case <-time.After(opts.IdleTimeout):
			fmt.Printf("Idle timeout hit: %s\n", opts.IdleTimeout)
			return nil

		case event := <-watcher.Events:
			fmt.Printf("Event: %s", event)

			if isNewDir(event, runner.excludes) {
				fmt.Printf("Watching new directory: %s\n", event.Name)
				watcher.Add(event.Name)
				continue
			}
			runner.HandleEvent(event)

		case err := <-watcher.Errors:
			return err
		}
	}
}

func isNewDir(event fsnotify.Event, exclude *files.ExcludeList) bool {
	if event.Op&fsnotify.Create != fsnotify.Create {
		return false
	}

	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		fmt.Printf("Failed to stat %s: %s\n", event.Name, err)
		return false
	}

	return fileInfo.IsDir() && !exclude.IsMatch(event.Name)
}
