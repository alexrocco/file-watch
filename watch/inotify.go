package watch

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
)

type inotifyWatcher struct {
	log *logrus.Logger
}

// NewInotifyWatcher creates a Watcher using inotify
func NewInotifyWatcher(log *logrus.Logger) *inotifyWatcher {
	return &inotifyWatcher{log}
}

func (i *inotifyWatcher) Watches(dir string, fileModified chan FileModified, watchStarted chan bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "error creating the fsnotify watcher")
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				fInfo, err := os.Stat(event.Name)
				if err != nil {
					i.log.Error(errors.Wrapf(err, "error stating %q", event.Name))
				}

				// Ignore changes on directories
				if !fInfo.IsDir() {
					// Select the file operation
					var fileOp FileOperation
					switch event.Op {
					case fsnotify.Write:
						fileOp = Write
					case fsnotify.Remove:
						fileOp = Remove
					case fsnotify.Create:
						fileOp = Create
					case fsnotify.Rename:
						fileOp = Rename
					}

					// Push the modified file to the channel
					fileModified <- FileModified{
						Path:          event.Name,
						FileOperation: fileOp,
					}

					i.log.Debugf("Inotify watch event: file %q, file operation: %q", event.Name, fileOp)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				i.log.Error(err)
			}
		}
	}()

	// Add watcher to all the directories, including the main directory
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return errors.Wrapf(err, "error adding watcher on %q", path)
			}
		}

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "error adding sub directories in watcherfor %q", dir)
	}

	watchStarted <- true
	<-done

	return nil
}
