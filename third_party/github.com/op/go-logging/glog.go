package logging

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// syncBuffer joins a bufio.Writer to its underlying file, providing access to the
// file's Sync method and providing a wrapper for the Write method that provides log
// file rotation. There are conflicting methods, so the file cannot be embedded.
// l.mu is held for all its methods.
type syncBuffer struct {
	mu sync.Mutex
	*bufio.Writer
	file   *os.File
	nbytes uint64 // The number of bytes written to this file
}

var logging *syncBuffer

// createFiles creates all the log files for severity from sev down to infoLog.
func createFiles() *syncBuffer {
	now := time.Now()
	sb := &syncBuffer{}
	if err := sb.rotateFile(now); err != nil {
		panic("createFiles: " + err.Error())
	}
	logging = sb
	go sb.flushDaemon()
	return sb
}

func (sb *syncBuffer) Write(p []byte) error {
	sb.mu.Lock()
	if sb.nbytes+uint64(len(p)) >= MaxSize {
		if err := sb.rotateFile(time.Now()); err != nil {
			sb.flushAll()
			os.Exit(2)
		}
	}
	n, err := sb.Writer.Write(p)
	sb.nbytes += uint64(n)
	if err != nil {
		sb.flushAll()
		os.Exit(2)
	}
	sb.mu.Unlock()
	return err
}

// bufferSize sizes the buffer associated with each log file. It's large
// so that log records can accumulate without the logging thread blocking
// on disk I/O. The flushDaemon will block instead.
const bufferSize = 256 * 1024

// rotateFile closes the syncBuffer's file and starts a new one.
func (sb *syncBuffer) rotateFile(now time.Time) error {
	if sb.file != nil {
		sb.Flush()
		sb.file.Close()
	}
	var err error
	sb.file, _, err = create("log", now)
	sb.nbytes = 0
	if err != nil {
		return err
	}

	sb.Writer = bufio.NewWriterSize(sb.file, bufferSize)

	// Write header.
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Log file created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Running on machine: %s\n", host)
	fmt.Fprintf(&buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg\n")
	n, err := sb.file.Write(buf.Bytes())
	sb.nbytes += uint64(n)
	return err
}

const flushInterval = 30 * time.Second

// flushDaemon periodically flushes the log file buffers.
func (sb *syncBuffer) flushDaemon() {
	for _ = range time.NewTicker(flushInterval).C {
		sb.lockAndFlushAll()
	}
}

// lockAndFlushAll is like flushAll but locks l.mu first.
func (sb *syncBuffer) lockAndFlushAll() {
	sb.mu.Lock()
	sb.flushAll()
	sb.mu.Unlock()
}

// flushAll flushes all the logs and attempts to "sync" their data to disk.
// l.mu is held.
func (sb *syncBuffer) flushAll() {
	// Flush from fatal down, in case there's trouble flushing.
	if sb.file != nil {
		sb.Flush()     // ignore error
		sb.file.Sync() // ignore error
	}
}

// Flush flushes all pending log I/O.
func Flush() {
	logging.lockAndFlushAll()
}
