package ioutil

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func ReadAll(rd *bufio.Reader, d []byte) (err error) {
	tl, n, t := len(d), 0, 0
	for {
		if t, err = rd.Read(d[n:]); err != nil {
			return
		}
		if n += t; n == tl {
			break
		}
	}
	return
}

func WritePidFile(filename string) (err error) {
	if err = os.MkdirAll(filepath.Dir(filename), os.FileMode(0755)); err != nil {
		return
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return
	}
	if _, err = fmt.Fprintf(f, "%d\n", os.Getpid()); err != nil {
		f.Close()
		return
	}
	if err = f.Close(); err != nil {
		return
	}
	return
}
