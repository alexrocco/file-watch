package watch

// FileOperation the action that changed the file
type FileOperation int

const (
	Create FileOperation = iota
	Write
	Remove
	Rename
)

// FileModified describe a modification in a file
type FileModified struct {
	Path          string
	FileOperation FileOperation
}

// Watcher watches for changes in a specific path
type Watcher interface {
	// Watches watches for changes in the directory, including sub directories, and pushes them to the channel
	Watches(dir string, fileModified chan FileModified) error
}
