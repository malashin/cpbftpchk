package xftp

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// IFtp -
type IFtp interface {
	//Close()
	FileSize(path string) (int64, error)
	Delete(path string) error
	Rename(from, to string) error
	StorFrom(path string, r io.Reader, offset uint64) error
	List(path string) ([]TEntry, error)
	Quit() error
}

// TEntry -
type TEntry struct {
	Name string
	Size int64
	Type int
}

// -
const (
	File = iota
	Folder
	Unknown
)

// New -
// conn: user:pswd@proto://host/path:port
// proto - currently ftp or sftp only
// all fields are necessary
func New(conn string) (IFtp, error) {
	cs, err := ParseConnString(conn)
	if err != nil {
		return nil, err
	}
	addr := cs.Host + ":" + cs.Port
	switch cs.Proto {
	default:
		return nil, fmt.Errorf("unsupported protocol %q", cs.Proto)
	case "ftp":
		conn, err := ftp.DialTimeout(addr, 5*time.Second)
		if err != nil {
			return nil, err
		}
		err = conn.Login(cs.User, cs.Password)
		if err != nil {
			return nil, err
		}
		c := &TFtp{}
		c.client = conn
		return c, nil
	case "sftp":
		config := &ssh.ClientConfig{
			User: cs.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(cs.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			// Config: ssh.Config{
			// 	//Ciphers: []string{"aes128-cbc"},
			// 	Ciphers: []string{"3des-cbc", "aes256-cbc", "aes192-cbc", "aes128-cbc"},
			// },
		}
		conn, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			return nil, fmt.Errorf("Failed to dial: " + err.Error())
		}
		client, err := sftp.NewClient(conn)
		if err != nil {
			return nil, fmt.Errorf("Failed to create client: " + err.Error())
		}
		c := &TSftp{}
		c.client = client
		// c.client.
		return c, nil
	}
}

// TConnStruct -
type TConnStruct struct {
	User, Password, Proto, Host, Path, Port string
}

// ParseConnString -
func ParseConnString(conn string) (*TConnStruct, error) {
	const (
		user = iota
		password
		host
		port
		proto
		path
	)
	err := fmt.Errorf("wrong format")
	cs := TConnStruct{}
	parts := strings.Split(conn, ":")
	if len(parts) < 3 {
		return nil, err
	}

	idx := strings.LastIndex(parts[password], "@")
	if idx < 0 {
		return nil, err
	}
	parts = append(parts, parts[password][idx+1:])
	parts[password] = parts[password][:idx]

	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if !strings.HasPrefix(parts[host], "//") {
		return nil, err
	}
	parts[host] = parts[host][2:]
	if len(parts[host]) == 0 {
		return nil, err
	}
	if parts[port] == "" {
		parts[port] = "21"
	}
	if _, err := strconv.Atoi(parts[port]); err != nil {
		return nil, err
	}

	idx = strings.Index(parts[host], "/")
	if idx < 0 {
		return nil, err
	}
	parts = append(parts, parts[host][idx+1:])
	parts[host] = parts[host][:idx]

	cs.User = parts[user]
	cs.Password = parts[password]
	cs.Host = parts[host]
	cs.Path = parts[path]
	cs.Port = parts[port]
	cs.Proto = parts[proto]
	return &cs, nil
}
