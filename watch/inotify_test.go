package watch

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func Test_inotifyWatcher_Watches(t *testing.T) {
	t.Run("It should watch for files in the main directory", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(os.TempDir(), "")
		assert.NoError(t, err)

		log := logrus.New()
		log.Out = ioutil.Discard

		inotifyWatcher := NewInotifyWatcher(log)

		fileModified := make(chan FileModified)
		watchStarted := make(chan bool)

		go func() {
			err = inotifyWatcher.Watches(tmpDir, fileModified, watchStarted)
			assert.NoError(t, err)
		}()

		// Wait for the watch to start for starting the test
		<-watchStarted

		done := make(chan bool)
		go func() {
			tmpFile, err := os.CreateTemp(tmpDir, "")
			assert.NoError(t, err)

			expectedFileModified := FileModified{
				Path:          tmpFile.Name(),
				FileOperation: Create,
			}

			for {
				select {
				case gotFileModified := <-fileModified:
					assert.Equal(t, expectedFileModified, gotFileModified)
					done <- true
					return
				}
			}
		}()

		<-done

		// Remove the test files
		_ = os.RemoveAll(tmpDir)
	})
	t.Run("It should watch for files in the sub directory", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(os.TempDir(), "")
		subTmpDir, err := ioutil.TempDir(tmpDir, "")
		assert.NoError(t, err)

		log := logrus.New()
		log.Out = ioutil.Discard

		inotifyWatcher := NewInotifyWatcher(log)

		fileModified := make(chan FileModified)
		watchStarted := make(chan bool)

		go func() {
			err = inotifyWatcher.Watches(tmpDir, fileModified, watchStarted)
			assert.NoError(t, err)
		}()

		// Wait for the watch to start for starting the test
		<-watchStarted

		done := make(chan bool)
		go func() {
			tmpFile, err := os.CreateTemp(subTmpDir, "")
			assert.NoError(t, err)

			expectedFileModified := FileModified{
				Path:          tmpFile.Name(),
				FileOperation: Create,
			}

			for {
				select {
				case gotFileModified := <-fileModified:
					assert.Equal(t, expectedFileModified, gotFileModified)
					done <- true
					return
				}
			}
		}()

		<-done

		// Remove the test files
		_ = os.RemoveAll(tmpDir)
	})
}
