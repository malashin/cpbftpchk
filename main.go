package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github.com/macroblock/rawin"

	"github.com/macroblock/cpbftpchk/xftp"
	"github.com/macroblock/zl/core/zlog"
	"github.com/macroblock/zl/core/zlogger"
)

var (
	log = zlog.Instance("main")

	ftp        = xftp.IFtp(nil)
	remoteList = []xftp.TEntry{}
	args       = ""
)

func printStat(opt *xftp.TConnStruct) {
	log.Warning(true, "<ctrl-q> quit | <ctrl-r> refresh | <ctrl-s> paste")
	log.Info(fmt.Sprintf("[%v] %v/%v:%v", opt.Proto, opt.Host, opt.Path, opt.Port))
	log.Info("-------------------------------------------------")
}

func lookUpFile(name string, list []xftp.TEntry) bool {
	for _, file := range list {
		if name == file.Name {
			return true
		}
	}
	return false
}

func reloadList(path string) {
	log.Info("reading remote directory...")
	list, err := ftp.List(path)
	if err != nil {
		log.Warning(true, "lost connection")
		log.Info("reconnecting...")
		log.Warning(ftp.Quit(), "ftp.Quit()")
		ftp, err = xftp.New(args)
		if err != nil {
			log.Error(err, "xftp.New()")
			return
		}
		log.Info("reading remote directory...")
		list, err = ftp.List(path)
		if err != nil {
			log.Error(err, "ftp.List()")
			return
		}
	}
	remoteList = list
}

func process(ftp xftp.IFtp, opt *xftp.TConnStruct) {
	if ftp == nil {
		return
	}
	text, err := clipboard.ReadAll()
	if err != nil {
		log.Error(err, "clipboard.ReadAll()")
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		idx := strings.IndexFunc(line, func(r rune) bool {
			return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '.'
		})
		if idx < 0 {
			continue
		}
		line = line[idx:]
		idx = strings.IndexFunc(line, func(r rune) bool {
			return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '.')
		})
		if idx < 0 {
			idx = len(line)
		}
		line = line[:idx]
		if line == "" {
			continue
		}
		if lookUpFile(line, remoteList) {
			log.Notice("+ ", line)
			continue
		}
		if lookUpFile(line+".part", remoteList) {
			log.Warning(true, "? ", line)
			continue
		}
		log.Error(true, "- ", line)
	}
}

func main() {
	quit := false
	busy := true

	log.Add(zlogger.Build().Format("~x~e\n").Styler(zlogger.AnsiStyler).Done())

	if len(os.Args) > 1 {
		args = os.Args[1]
	}
	opt, err := xftp.ParseConnString(args)

	err = rawin.Start()
	defer rawin.Stop()
	if err != nil {
		log.Error(err, "start console raw mode")
		return
	}

	rawin.SetAction(rawin.PreFilter, func(r rune) bool { fmt.Printf("%q %U\n", r, r); return false })
	// ctrl-q
	rawin.SetAction('\x11', func(r rune) bool {
		quit = true
		return true
	})
	// ctrl-r
	rawin.SetAction('\x12', func(r rune) bool {
		if !busy {
			busy = true
			log.Info("-------------------------------------------------")
			reloadList(opt.Path)
			printStat(opt)
			busy = false
		}
		return true
	})
	// ctrl-s
	rawin.SetAction('\x13', func(r rune) bool {
		if !busy {
			busy = true
			process(ftp, opt)
			log.Info("-------------------------------------------------")
			printStat(opt)
			busy = false
		}
		return true
	})

	log.Info("connecting...")
	ftp, err = xftp.New(args)
	if err != nil {
		log.Error(err, "xftp.New()")
		log.Warning(true, "format:")
		log.Warning(true, "    user:pswd@proto://host/path:port")
		return
	}
	defer ftp.Quit()

	reloadList(opt.Path)
	printStat(opt)

	busy = false
	for !quit {
		time.Sleep(50 * time.Millisecond)
	}
}
