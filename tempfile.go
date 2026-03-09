package tempfile

import (
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/rclancey/ensuredir"
)

type TempFile struct {
	*os.File
	realFn string
	tempFn string
	hash hash.Hash
}

type Option func(tf *TempFile) error

func WithHash(h hash.Hash) Option {
	return func(tf *TempFile) error {
		tf.hash = h
		return nil
	}
}

func Create(fn string, opts ...Option) (*TempFile, error) {
	err := ensuredir.EnsureDir(filepath.Dir(fn))
	if err != nil {
		return nil, err
	}
	f, err := os.CreateTemp(filepath.Dir(fn), "."+filepath.Base(fn)+"*.tmp")
	if err != nil {
		return nil, err
	}
	tf := &TempFile{
		File: f,
		realFn: fn,
		tempFn: f.Name(),
	}
	for _, opt := range opts {
		err = opt(tf)
		if err != nil {
			tf.Abandon()
			return nil, err
		}
	}
	return tf, nil
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

func (tf *TempFile) Write(buf []byte) (int, error) {
	n, err := tf.File.Write(buf)
	if tf.hash != nil {
		tf.hash.Write(buf[:n])
	}
	return n, err
}

func (tf *TempFile) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 65536)
	var written int64
	for {
		n, err := r.Read(buf)
		if n > 0 {
			wn, werr := tf.Write(buf[:n])
			written += int64(wn)
			if werr != nil {
				return written, werr
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return written, nil
			}
			return written, err
		}
	}
}

func (tf *TempFile) Hash() []byte {
	if tf.hash == nil {
		return nil
	}
	return tf.hash.Sum(nil)
}

func (tf *TempFile) HexHash() string {
	return hex.EncodeToString(tf.Hash())
}
