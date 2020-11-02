package xftp

import (
	"io"

	"github.com/malashin/ftp"
)

// TFtp -
type TFtp struct {
	client *ftp.ServerConn
}

// Delete -
func (o *TFtp) Delete(path string) error {
	return o.client.Delete(path)
}

// Exists -
func (o *TFtp) Exists(path string) error {
	_, err := o.FileSize(path)
	if err != nil {
		return err
	}
	return nil
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

// RetrFrom issues a RETR FTP command to fetch the specified file from the remote
// FTP server, the server will not send the offset first bytes of the file.
func (o *TFtp) RetrFrom(path string, w io.Writer, offset uint64) error {
	r, err := o.client.RetrFrom(path, offset)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	r.Close()
	return nil
}

// ChangeDir -
func (o *TFtp) ChangeDir(dir string) error {
	return o.client.ChangeDir(dir)
}

// ChangeDirToParent -
func (o *TFtp) ChangeDirToParent() error {
	return o.client.ChangeDirToParent()
}

// CurrentDir -
func (o *TFtp) CurrentDir() (string, error) {
	return o.client.CurrentDir()
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
