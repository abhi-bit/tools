package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type appLogCloser struct {
	path      string
	filePtr   unsafe.Pointer //Stores file pointer
	perm      os.FileMode
	maxSize   int64
	maxFiles  int64
	size      int64
	lowIndex  int64
	highIndex int64
	exitCh    chan bool
	mu        sync.Mutex
	once      sync.Once
}

func (wc *appLogCloser) Write(p []byte) (_ int, err error) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.once.Do(func() {
		wc.exitCh = make(chan bool, 1)
		go wc.cleanupTask()
	})
	fptr := (*os.File)(atomic.LoadPointer(&wc.filePtr))
	if fptr == nil {
		return 0, err
	}
	bytesWritten, err := fptr.Write(p)
	atomic.AddInt64(&wc.size, int64(bytesWritten))
	return bytesWritten, err
}

func (wc *appLogCloser) Close() error {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	fptr := (*os.File)(atomic.LoadPointer(&wc.filePtr))
	wc.exitCh <- true
	if fptr != nil {
		return fptr.Close()
	}
	return nil
}

func (wc *appLogCloser) manageLogFiles() {
	if err := os.Rename(wc.path, fmt.Sprintf("%s.%d", wc.path, wc.highIndex+1)); err != nil {
		return
	}
	wc.highIndex += 1
	file, err := os.OpenFile(wc.path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	old := atomic.LoadPointer(&wc.filePtr)
	if atomic.CompareAndSwapPointer(&wc.filePtr, old, unsafe.Pointer(file)) == true {
		atomic.StoreInt64(&wc.size, 0)
		if old != nil {
			_ = (*os.File)(old).Close()
		}
	}
	for ; wc.lowIndex+wc.maxFiles <= wc.highIndex; wc.lowIndex++ {
		_ = os.Remove(fmt.Sprintf("%s.%d", wc.path, wc.lowIndex))
	}
}

func (wc *appLogCloser) cleanupTask() {
	for {
		select {
		case <-wc.exitCh:
			return
		default:
		}
		if wc.maxSize <= atomic.LoadInt64(&wc.size) {
			wc.manageLogFiles()
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func getFileIndexRange(path string) (int64, int64) {
	files, err := filepath.Glob(path + ".*")
	if err != nil || len(files) == 0 {
		return 1, 0
	}
	var lowIndex int64 = (1 << 63) - 1
	var highIndex int64 = 0
	for _, file := range files {
		tokens := strings.Split(file, ".")
		if index, err := strconv.ParseInt(tokens[len(tokens)-1], 10, 64); err == nil {
			if index < lowIndex {
				lowIndex = index
			}

			if index > highIndex {
				highIndex = index
			}
		}
	}
	return lowIndex, highIndex
}

func OpenAppLog(path string, perm os.FileMode, maxSize, maxFiles int64) (io.WriteCloser, error) {
	if maxSize < 1 {
		return nil, fmt.Errorf("maxSize should be > 1")
	}
	if maxFiles < 1 {
		return nil, fmt.Errorf("maxFiles should be > 1")
	}

	// If path exists determine size and check path is a regular file.
	var size int64
	fi, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		if fi.Mode()&os.ModeType != 0 {
			return nil, fmt.Errorf("Supplied app log file, path: %s is not a regular file", path)
		}
		size = fi.Size()
	}

	// Open path for reading/writing, create if necessary.
	file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return nil, err
	}

	low, high := getFileIndexRange(path)

	return &appLogCloser{
		path:      path,
		filePtr:   unsafe.Pointer(file),
		perm:      perm,
		maxSize:   maxSize,
		maxFiles:  maxFiles,
		size:      size,
		lowIndex:  low,
		highIndex: high,
	}, nil
}

func updateApplogSetting(wc *appLogCloser, maxFileCount, maxFileSize int64) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.maxFiles = maxFileCount
	wc.maxSize = maxFileSize
}
