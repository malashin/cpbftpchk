package xftp

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/macroblock/ptool"
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
	Time time.Time
	Type int
}

// -
const (
	File = iota
	Folder
	Unknown
)

var urlParser *ptool.TParser

func init() {
	//[proto://][username[:password]@]host[/path][:port]
	urlRule := "" +
		"entry = spaces [@proto '://'] [@username[':'@password]'@'] @host ['/'@path] ['/'] [':'@port] spaces$;" +
		"spaces   = {'\x00'..'\x20'};" +
		"anyRune  = !$ '\x00'..'\xff';" +
		"invalid  = '\x00'..'\x20';" +
		"digit    = '0'..'9';" +
		"proto    = {!invalid!':' anyRune};" +
		"username = {!invalid!':'!'@' anyRune};" +
		"password = {!invalid!'@' anyRune};" +
		"host     = {!invalid!':'!'/' anyRune};" +
		"path     = {!invalid!':'!'/:'!('/'$) anyRune};" +
		"port     = digit{digit};"

	rules := ptool.NewRules(urlRule)
	p, err := rules.Parser()
	if err != nil {
		log.Panic(err, "can't compile url parser")
		panic(err)
	}
	urlParser = p
}

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
		err = conn.Login(cs.Username, cs.Password)
		if err != nil {
			return nil, err
		}
		c := &TFtp{}
		c.client = conn
		return c, nil
	case "sftp":
		config := &ssh.ClientConfig{
			User: cs.Username,
			Auth: []ssh.AuthMethod{
				ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
					// Just send the password back for all questions
					answers := make([]string, len(questions))
					for i := range answers {
						answers[i] = cs.Password
					}
					return answers, nil
				}),
				ssh.PasswordCallback(func() (string, error) { return cs.Password, nil }),
				ssh.Password(cs.Password),
			},
			// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: 10 * time.Second,
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

func setDefault(val string, def string) string {
	if val == "" {
		return def
	}
	return val
}

// TConnStruct -
type TConnStruct struct {
	Proto, Username, Password, Host, Path, Port string
}

// ParseConnString -
func ParseConnString(conn string) (*TConnStruct, error) {
	cs := TConnStruct{}
	tree, err := urlParser.Parse(conn)
	if err != nil {
		return nil, err
	}
	for _, node := range tree.Links {
		switch urlParser.ByID(node.Type) {
		case "proto":
			cs.Proto = node.Value
		case "username":
			cs.Username = node.Value
		case "password":
			cs.Password = node.Value
		case "host":
			cs.Host = node.Value
		case "path":
			cs.Path = node.Value
		case "port":
			cs.Port = node.Value
		}
	}
	// fmt.Printf("proto %q\nuser %q\npswd %q\nhost %q\npath %q\nport %q\n", cs.Proto, cs.Username, cs.Password, cs.Host, cs.Path, cs.Port)
	if cs.Port == "" {
		switch cs.Proto {
		case "":
			cs.Proto = "ftp"
			cs.Port = "21"
		case "ftp":
			cs.Port = "21"
		case "sftp":
			cs.Port = "22"
		}
	}
	if cs.Host == "" || cs.Path == "" {
		return nil, fmt.Errorf("no host or path name")
	}
	return &cs, nil
	// const (
	// 	user = iota
	// 	password
	// 	host
	// 	port
	// 	proto
	// 	path
	// )
	// err := fmt.Errorf("wrong format")
	// cs := TConnStruct{}
	// parts := strings.Split(conn, ":")
	// if len(parts) < 3 {
	// 	return nil, err
	// }
	// if len(parts) < 4 {
	// 	parts = append(parts, "")
	// }

	// idx := strings.LastIndex(parts[password], "@")
	// if idx < 0 {
	// 	return nil, err
	// }
	// parts = append(parts, parts[password][idx+1:])
	// parts[password] = parts[password][:idx]

	// for i := range parts {
	// 	parts[i] = strings.TrimSpace(parts[i])
	// }
	// if !strings.HasPrefix(parts[host], "//") {
	// 	return nil, err
	// }
	// parts[host] = parts[host][2:]
	// if len(parts[host]) == 0 {
	// 	return nil, err
	// }
	// if parts[port] == "" {
	// 	parts[port] = "21"
	// }
	// if _, err := strconv.Atoi(parts[port]); err != nil {
	// 	return nil, err
	// }

	// idx = strings.Index(parts[host], "/")
	// if idx < 0 {
	// 	return nil, err
	// }
	// parts = append(parts, parts[host][idx+1:])
	// parts[host] = parts[host][:idx]

	// cs.User = parts[user]
	// cs.Password = parts[password]
	// cs.Host = parts[host]
	// cs.Path = parts[path]
	// cs.Port = parts[port]
	// cs.Proto = parts[proto]
	// return &cs, nil
}
