package main

// #include <time.h>
// #include <lastlog.h>
import "C"

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

// lastlog.h struct lastlog
var llsize = C.sizeof_struct_lastlog
var ttsize = C.sizeof_time_t

// version by Makefile
var version string

type cmdOpts struct {
	Before         int64  `long:"before" description:"[Deprecated] Check for users whose login is older than DAYS"`
	Warn           int64  `short:"w" long:"warning" default:"60" description:"warning if users whose login is older than DAYS"`
	Crit           int64  `short:"c" long:"critical" default:"85" description:"critical if users whose login is older than DAYS"`
	MinUID         int    `long:"min-uid" default:"500" description:"min uid to check lastlog"`
	MaxUID         int    `long:"max-uid" default:"60000" description:"max uid to check lastlog"`
	WhiteUserNames string `long:"white-user-names" default:"" description:"comma separeted user names that white"`
	Version        bool   `short:"v" long:"version" description:"Show version"`
}

// User :
type User struct {
	UID      int
	UserName string
	Shell    string
	LastLog  int64
}

// LastLogTime : user.LastLog as time.TIme
func (u *User) LastLogTime() time.Time {
	return time.Unix(u.LastLog, 0)
}

var noLoginShell = map[string]struct{}{
	"/bin/sync":      struct{}{},
	"/sbin/halt":     struct{}{},
	"/sbin/nologin":  struct{}{},
	"/sbin/shutdown": struct{}{},
}

// NoLogin : User has nologin shell
func (u *User) NoLogin() bool {
	_, ok := noLoginShell[u.Shell]
	return ok
}

// LastLoginDays :
func (u *User) LastLoginDays() string {
	if u.LastLog == 0 {
		return "*Never logged in*"
	}
	t := time.Now().Unix() - u.LastLog
	if t < 0 {
		t = 0
	}
	return fmt.Sprintf("%d days", int(t/86400))
}

func readLastLog() (map[int]int64, error) {
	lastlog := make(map[int]int64)
	f, err := os.Open("/var/log/lastlog")
	if err != nil {
		return lastlog, err
	}
	defer f.Close()
	buf := make([]byte, llsize)
	pos := 0
	for {
		n, err := f.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			return lastlog, err
		}
		unixTime := int64(binary.LittleEndian.Uint32(buf[:ttsize]))
		lastlog[pos] = unixTime
		pos++
	}
	return lastlog, nil
}

func readPasswd() ([]User, error) {
	users := make([]User, 0)
	lastLog, err := readLastLog()
	if err != nil {
		return users, err
	}

	f, err := os.Open("/etc/passwd")
	if err != nil {
		return users, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		// kevin:x:1005:1006::/home/kevin:/usr/bin/zsh
		parts := strings.SplitN(line, ":", 7)
		if len(parts) < 6 || parts[0] == "" ||
			parts[0][0] == '+' || parts[0][0] == '-' {
			continue
		}
		uid, err := strconv.Atoi(parts[2])
		if err != nil {
			return users, err
		}
		ll, ok := lastLog[uid]
		if !ok {
			ll = 0
		}
		u := User{
			UID:      uid,
			UserName: parts[0],
			Shell:    parts[6],
			LastLog:  ll,
		}
		users = append(users, u)
	}
	return users, nil
}

func checkLastLog(opts cmdOpts) (int, string) {
	whileUserNames := make(map[string]struct{})
	if opts.WhiteUserNames != "" {
		names := strings.Split(opts.WhiteUserNames, ",")
		for _, n := range names {
			whileUserNames[n] = struct{}{}
		}
	}

	now := time.Now().Unix()
	warn := now - opts.Warn*86400
	crit := now - opts.Crit*86400

	// for compatibility
	if opts.Before != 0 {
		warn = now - opts.Before*86400
		crit = now - opts.Before*86400
	}

	chkSt := 0
	msgs := make([]string, 0)

	users, err := readPasswd()
	if err != nil {
		return -1, fmt.Sprintf("UNKNOWN: %v", err)
	}
	for _, u := range users {
		if u.UID <= opts.MinUID {
			continue
		}
		if u.UID >= opts.MaxUID {
			continue
		}
		if _, ok := whileUserNames[u.UserName]; ok {
			continue
		}
		if u.NoLogin() {
			continue
		}

		if u.LastLog < crit {
			chkSt |= 2
			msgs = append(msgs, fmt.Sprintf("%s(%s)", u.UserName, u.LastLoginDays()))
		} else if u.LastLog < warn {
			chkSt |= 1
			msgs = append(msgs, fmt.Sprintf("%s(%s)", u.UserName, u.LastLoginDays()))
		}
	}

	if (chkSt & 2) == 2 {
		// crit
		return 2, fmt.Sprintf("CRITICAL: Found users who have not logged in recently: %s", strings.Join(msgs, ", "))
	} else if (chkSt & 1) == 1 {
		// warn
		return 1, fmt.Sprintf("WARNING: Found users who have not logged in recently: %s", strings.Join(msgs, ", "))
	} else {
		// ok
		return 0, "OK: No users were found who have not logged in recently"
	}
}

func printVersion() {
	fmt.Printf(`%s %s
Compiler: %s %s
sizeof(lastlog): %d
`,
		os.Args[0],
		version,
		runtime.Compiler,
		runtime.Version(),
		llsize)
}

func main() {
	opts := cmdOpts{}
	psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opts.Version {
		printVersion()
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	chkSt, msg := checkLastLog(opts)
	fmt.Println(msg)
	os.Exit(chkSt)
}
