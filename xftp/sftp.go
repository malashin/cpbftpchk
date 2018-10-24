package xftp

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
)

// TSftp -
type TSftp struct {
	client *sftp.Client
	cwd    string
}

func (o *TSftp) resolveDir(dir string) (string, error) {
	if path.IsAbs(dir) {
		return path.Clean(dir), nil
	}
	if o.cwd != "" && path.IsAbs(o.cwd) {
		return path.Join(o.cwd, dir), nil
	}
	base, err := o.client.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(base, o.cwd, dir), nil
}

// Delete -
func (o *TSftp) Delete(path string) error {
	return o.client.Remove(path)
}

// Rename -
func (o *TSftp) Rename(from, to string) error {
	return o.client.Rename(from, to)
}

// Quit -
func (o *TSftp) Quit() error {
	return o.client.Close()
}

// FileSize -
func (o *TSftp) FileSize(path string) (int64, error) {
	stat, err := o.client.Stat(path)
	if err != nil {
		return -1, err
	}
	return stat.Size(), nil
}

// StorFrom -
func (o *TSftp) StorFrom(path string, r io.Reader, offset uint64) error {
	// conn, err := c.cmdDataConnFrom(offset, "STOR %s", path)
	// if err != nil {
	// 	return err
	// }
	var err error
	path, err = o.resolveDir(path)
	if err != nil {
		return err
	}
	f, err := o.client.OpenFile(path, os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer f.Close()
	//_, err = io.Copy(f, r)
	offs, err := f.Seek(int64(offset), 0)
	if err != nil {
		return err
	}
	if offs != int64(offset) {
		return fmt.Errorf("Sftp Seek() problem (custom error). Search %v, but return %v", int64(offset), offs)
	}
	_, err = f.ReadFrom(r)
	if err != nil {
		return err
	}
	// _, _, err = c.conn.ReadResponse(StatusClosingDataConnection)
	// return err
	return nil
}

// ChangeDir -
func (o *TSftp) ChangeDir(dir string) error {
	var err error
	o.cwd, err = o.resolveDir(dir)
	if err != nil {
		return err
	}
	return nil
}

// CurrentDir -
func (o *TSftp) CurrentDir() (string, error) {
	return o.cwd, nil
}

// List -
func (o *TSftp) List(path string) ([]TEntry, error) {
	var err error
	path, err = o.resolveDir(path)
	if err != nil {
		return nil, err
	}
	src, err := o.client.ReadDir(path)
	if err != nil {
		return nil, err
	}
	list := []TEntry{}
	for _, item := range src {
		entry := TEntry{Name: item.Name(), Size: item.Size(), Time: item.ModTime()}
		entry.Type = File
		if item.IsDir() {
			entry.Type = Folder

		}
		list = append(list, entry)
	}
	return list, nil
}
