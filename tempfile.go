package tempfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ensureDir(dn string) error {
	st, err := os.Stat(dn)
	if err == nil {
		if !st.IsDir() {
			return fmt.Errorf("%w: %s exists and is not a directory", os.ErrExist, dn)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("can't stat %s: %w", dn, err)
	}
	err = os.MkdirAll(dn, 0775)
	if err != nil {
		return fmt.Errorf("can't mkdir %s: %w", dn, err)
	}
	return nil
}

type TempFile struct {
	*os.File
	realFn string
	tempFn string
}

func Create(fn string) (*TempFile, error) {
	err := ensureDir(filepath.Dir(fn))
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
