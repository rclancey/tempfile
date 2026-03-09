package tempfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rclancey/ensuredir"
)

type TempFile struct {
	*os.File
	realFn string
	tempFn string
}

func Create(fn string) (*TempFile, error) {
	err := ensuredir.EnsureDir(filepath.Dir(fn))
	if err != nil {
		return nil, err
	}
	tf, err := os.CreateTemp(filepath.Dir(fn), "."+filepath.Base(fn)+"*.tmp")
	if err != nil {
		return nil, err
	}
	return &TempFile{
		File: tf,
		realFn: fn,
		tempFn: tf.Name(),
	}, nil
}

func (tf *TempFile) Abandon() error {
	tf.File.Close()
	return os.Remove(tf.tempFn)
}

func (tf *TempFile) Close() error {
	err := tf.File.Close()
	if err != nil {
		os.Remove(tf.tempFn)
		return err
	}
	err = os.Rename(tf.tempFn, tf.realFn)
	if err != nil {
		os.Remove(tf.tempFn)
		return err
	}
	return nil
}
