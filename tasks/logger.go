package tasks

import (
	"io"
	"log"
	"os"
	"path"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type stdlogger struct {
	logger *log.Logger
}

type fileLogger struct {
	stdlogger
	pathfile    string
	initialized bool
	file        *os.File
}

var _ Logger = &fileLogger{}

func StdLogger(w io.Writer) Logger {
	return &stdlogger{logger: log.New(w, "", log.LstdFlags)}
}

func FileLogger(pathfile string) *fileLogger {
	return &fileLogger{pathfile: pathfile}
}

func (f *fileLogger) Init() error {
	if f.initialized {
		return nil
	}
	dir := path.Dir(f.pathfile)
	if e, err := isExists(dir); err != nil {
		return err
	} else if !e {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	var err error
	if f.file, err = os.OpenFile(f.pathfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return err
	}
	f.stdlogger.logger = log.New(f.file, "", log.LstdFlags)
	return nil
}

func (f *stdlogger) Infof(format string, args ...interface{}) {
	f.logger.Printf(" INFO: "+format, args...)
}

func (f *stdlogger) Errorf(format string, args ...interface{}) {
	f.logger.Printf(" ERROR: "+format, args...)
}

func (f *fileLogger) Close() error {
	defer func() {
		f.stdlogger.logger = nil
	}()
	if f.initialized {
		return f.file.Close()
	}
	return nil
}
