package tasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

type Savepoint interface {
	SetOffset(ctx context.Context, offset int) error
	Offset(ctx context.Context, offset *int) error
}

type fileSavepoint struct {
	pathfile    string
	initialized bool
	file        *os.File
}

func FileSavepoint(pathfile string) *fileSavepoint {
	return &fileSavepoint{pathfile: pathfile}
}

func (f *fileSavepoint) Init() error {
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
	if f.file, err = os.OpenFile(f.pathfile, os.O_RDWR|os.O_CREATE, 0666); err != nil {
		return err
	}
	return nil
}

func (f *fileSavepoint) Close() error {
	if f.initialized {
		return f.file.Close()
	}
	return nil
}

func (savePoint *fileSavepoint) SetOffset(ctx context.Context, offset int) error {
	_, err := savePoint.file.WriteAt([]byte(fmt.Sprintf("%d\n", offset)), 0)
	return err
}

func (savePoint *fileSavepoint) Offset(ctx context.Context, offset *int) error {
	if _, err := savePoint.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	bs := make([]byte, 64)
	var err error
	if _, err = savePoint.file.Read(bs); err != nil {
		return err
	}
	idx := bytes.Index(bs, []byte{'\n'})
	if idx < 0 {
		return errors.New("save point format error")
	}
	*offset, err = strconv.Atoi(string(bs[:idx]))
	return err
}
