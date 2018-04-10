package xftp

import (
	"io"

	"github.com/jlaffaye/ftp"
)

// TFtp -
type TFtp struct {
	client *ftp.ServerConn
}

// Delete -
func (o *TFtp) Delete(path string) error {
	return o.client.Delete(path)
}

// Rename -
func (o *TFtp) Rename(from, to string) error {
	return o.client.Rename(from, to)
}

// Quit -
func (o *TFtp) Quit() error {
	return o.client.Quit()
}

// FileSize -
func (o *TFtp) FileSize(path string) (int64, error) {
	return o.client.FileSize(path)
}

// StorFrom -
func (o *TFtp) StorFrom(path string, r io.Reader, offset uint64) error {
	return o.client.StorFrom(path, r, offset)
}

// List -
func (o *TFtp) List(path string) ([]TEntry, error) {
	src, err := o.client.List(path)
	if err != nil {
		return nil, err
	}
	list := []TEntry{}
	for _, item := range src {
		entry := TEntry{Name: item.Name, Size: int64(item.Size), Time: item.Time}
		switch item.Type {
		default:
			entry.Type = Unknown
		case ftp.EntryTypeFile:
			entry.Type = File
		case ftp.EntryTypeFolder:
			entry.Type = Folder
		}
		list = append(list, entry)
	}
	return list, nil
}
