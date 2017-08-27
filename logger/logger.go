package logger

import (
	"os"
	"sync"
)

type Logger struct {
	path string
	mu   sync.Mutex
	file *os.File
}

func New(file string) (*Logger, error) {
	f, err := openFile(file)
	if err != nil {
		return nil, err
	}

	return &Logger{path: file, file: f}, nil
}

func (l *Logger) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	n, err := l.file.Write(p)
	if err != nil {
		return n, err
	}
	err = l.file.Sync()

	return n, err
}

func (l *Logger) Reopen() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	err := l.file.Close()
	if err != nil {
		return err
	}
	f, err := openFile(l.path)
	if err != nil {
		return err
	}
	l.file = f

	return nil
}

func openFile(file string) (*os.File, error) {
	return os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
