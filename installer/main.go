// Stickerz 1.0 installer for reMarkable (XOVI + qt-resource-rebuilder).
// Copies the stickerz package to the tablet and runs install.sh / uninstall.sh.
// Upgrades cleanly over rm-stickers v3 / v3.1.
// Does NOT restart the UI; you rebuild the hashtab and restart XOVI/xochitl in reManager.
package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

const (
	version     = "1.0.0"
	appName     = "Stickerz"
	defaultIP   = "10.11.99.1"
	defaultUsr  = "root"
	remoteRoot  = "/home/root/stickerz"
	xoviDir     = "/home/root/xovi"
	qmdDir      = xoviDir + "/exthome/qt-resource-rebuilder"
)

//go:embed all:payload
var payload embed.FS

var in = bufio.NewReader(os.Stdin)

func main() {
	code := run()
	fmt.Println()
	prompt("Press Enter to exit.")
	os.Exit(code)
}

func run() int {
	fmt.Printf("%s installer, version %s\n", appName, version)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("This installs, updates or removes Stickerz 1.0")
	fmt.Println("on a reMarkable tablet that already has XOVI and qt-resource-rebuilder.")
	fmt.Println()
	fmt.Println("Stickerz 1.0 is based on rm-stickers v3.1, which Nick confirmed works on his tablet.")
	fmt.Println("Read the README on the tablet (or in the package) before you install.")
	fmt.Println()
	fmt.Println("Before you start:")
	fmt.Println("  - Connect the tablet to this computer with its USB cable, switch it on and unlock it.")
	fmt.Println("  - You need the tablet's SSH password. On the tablet go to")
	fmt.Println("    Settings > General > Help > About > Copyrights and licenses and scroll down;")
	fmt.Println("    the password is shown there (developer mode is needed for SSH).")
	fmt.Println()

	host := ask("Tablet IP address", defaultIP)
	user := ask("Username", defaultUsr)
	pass := askPassword("Password (it won't show as you type): ")
	fmt.Println()

	fmt.Printf("Connecting to %s@%s ...\n", user, host)
	client, err := connect(host, user, pass)
	// Forget the password as soon as we have a connection (or have failed).
	pass = ""
	if err != nil {
		fmt.Println()
		fmt.Println(friendlyConnError(err, host))
		return 1
	}
	defer client.Close()
	fmt.Println("Connected.")

	if err := checkXOVI(client); err != nil {
		fmt.Println()
		fmt.Println(err)
		return 1
	}

	for {
		fmt.Println()
		fmt.Println("What would you like to do?")
		fmt.Println("  1  Install or update Stickerz 1.0")
		fmt.Println("  2  Uninstall Stickerz")
		fmt.Println("  3  Quit")
		switch strings.TrimSpace(prompt("Choose 1, 2 or 3: ")) {
		case "1":
			if err := install(client); err != nil {
				fmt.Println()
				fmt.Println("Installation failed:", err)
				fmt.Println("Nothing further was changed by this installer.")
				return 1
			}
			return 0
		case "2":
			if err := uninstall(client); err != nil {
				fmt.Println()
				fmt.Println("Uninstall failed:", err)
				return 1
			}
			return 0
		case "3", "q", "Q":
			fmt.Println("Nothing was changed.")
			return 0
		default:
			fmt.Println("Please type 1, 2 or 3.")
		}
	}
}

// ---------- prompts ----------

func prompt(msg string) string {
	fmt.Print(msg)
	line, err := in.ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}

func ask(label, def string) string {
	v := strings.TrimSpace(prompt(fmt.Sprintf("%s [%s]: ", label, def)))
	if v == "" {
		return def
	}
	return v
}

func askPassword(label string) string {
	fmt.Print(label)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		if err == nil {
			return string(b)
		}
	}
	// Not an interactive console (e.g. piped input): read a plain line.
	return prompt("")
}

func yesNo(q string, def bool) bool {
	d := "y/N"
	if def {
		d = "Y/n"
	}
	for {
		a := strings.ToLower(strings.TrimSpace(prompt(fmt.Sprintf("%s [%s]: ", q, d))))
		switch a {
		case "":
			return def
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
		fmt.Println("Please type y or n.")
	}
}

// ---------- SSH ----------

func sshConfig(user, pass string) *ssh.ClientConfig {
	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
			ssh.KeyboardInteractive(func(_, _ string, qs []string, _ []bool) ([]string, error) {
				ans := make([]string, len(qs))
				for i := range ans {
					ans[i] = pass
				}
				return ans, nil
			}),
		},
		// The tablet's host key changes with software updates, so accept whatever it presents.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	// Allow older algorithms too, in case the tablet runs an older Dropbear.
	sup, insec := ssh.SupportedAlgorithms(), ssh.InsecureAlgorithms()
	cfg.KeyExchanges = append(append([]string{}, sup.KeyExchanges...), insec.KeyExchanges...)
	cfg.Ciphers = append(append([]string{}, sup.Ciphers...), insec.Ciphers...)
	cfg.MACs = append(append([]string{}, sup.MACs...), insec.MACs...)
	cfg.HostKeyAlgorithms = append(append([]string{}, sup.HostKeys...), insec.HostKeys...)
	return cfg
}

func connect(host, user, pass string) (*ssh.Client, error) {
	addr := host
	if _, _, err := net.SplitHostPort(host); err != nil {
		addr = net.JoinHostPort(host, "22")
	}
	return ssh.Dial("tcp", addr, sshConfig(user, pass))
}

func friendlyConnError(err error, host string) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "unable to authenticate") || strings.Contains(msg, "no supported methods remain"):
		return "The tablet rejected the password (or username).\n" +
			"Check the password on the tablet under Settings > General > Help > About >\n" +
			"Copyrights and licenses. It is case-sensitive, and it can change after a software update.\n" +
			"The username is normally root."
	case strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "no route to host") ||
		strings.Contains(msg, "network is unreachable") || strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connectex") || strings.Contains(msg, "timed out"):
		return fmt.Sprintf("Couldn't reach the tablet at %s.\n"+
			"  - Check the USB cable is plugged in at both ends and the tablet is switched on and unlocked.\n"+
			"  - Over USB the address is normally 10.11.99.1.\n"+
			"  - If you're using Wi-Fi instead, make sure SSH over Wi-Fi is enabled and the address is right.\n"+
			"(Technical detail: %v)", host, err)
	case strings.Contains(msg, "no such host"):
		return fmt.Sprintf("'%s' isn't a valid address. Over USB the tablet is normally 10.11.99.1.", host)
	}
	return fmt.Sprintf("Couldn't connect to the tablet: %v", err)
}

// runCmd runs a shell command on the tablet and returns combined output.
func runCmd(c *ssh.Client, cmd string) (string, error) {
	s, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer s.Close()
	var buf lockedBuffer
	s.Stdout = &buf
	s.Stderr = &buf
	err = s.Run(cmd)
	return strings.TrimSpace(buf.String()), err
}

func checkXOVI(c *ssh.Client) error {
	out, err := runCmd(c, "if [ ! -d "+xoviDir+" ]; then echo NOXOVI; elif [ ! -d "+qmdDir+" ]; then echo NOQMD; else echo OK; fi")
	if err != nil {
		return fmt.Errorf("couldn't run a check on the tablet: %v %s", err, out)
	}
	switch out {
	case "OK":
		return nil
	case "NOXOVI":
		return errors.New("XOVI isn't installed on this tablet (" + xoviDir + " wasn't found).\n" +
			"Stickerz needs XOVI and qt-resource-rebuilder. Install those first, then run this again.\n" +
			"Nothing was changed.")
	default:
		return errors.New("XOVI is installed, but qt-resource-rebuilder isn't (" + qmdDir + " wasn't found).\n" +
			"Stickerz is a QMD mod for qt-resource-rebuilder. Install that extension first, then run this again.\n" +
			"Nothing was changed.")
	}
}

// ---------- install ----------

func install(c *ssh.Client) error {
	fmt.Println()
	fmt.Println("[1/3] Preparing " + remoteRoot + " on the tablet...")
	if out, err := runCmd(c, "rm -rf "+remoteRoot+" && mkdir -p "+remoteRoot+"/src "+remoteRoot+"/library"); err != nil {
		return fmt.Errorf("couldn't create folders: %v %s", err, out)
	}

	fmt.Println("[2/3] Copying the Stickerz 1.0 package to the tablet...")
	if err := uploadTree(c); err != nil {
		runCmd(c, "rm -rf "+remoteRoot)
		return err
	}

	fmt.Println("[3/3] Running install.sh on the tablet...")
	fmt.Println()
	// Feed a newline so a non-3.28 firmware warning's "read" does not hang under no TTY.
	out, err := runCmd(c, "cd "+remoteRoot+" && printf '\\n' | sh ./install.sh")
	if out != "" {
		fmt.Println(out)
		fmt.Println()
	}
	if err != nil {
		return fmt.Errorf("install.sh failed: %v", err)
	}

	fmt.Println("Success! Stickerz 1.0 files are on the tablet.")
	fmt.Println()
	fmt.Println("This installer does NOT restart the UI.")
	fmt.Println("Next steps — in reManager:")
	fmt.Println("  1. Rebuild the hashtab")
	fmt.Println("  2. Restart XOVI / xochitl")
	fmt.Println()
	fmt.Println("If xochitl crash-loops, the tablet reboots to stock")
	fmt.Println("(XOVI is tethered). Then SSH in, run uninstall from this installer (or")
	fmt.Println("sh " + remoteRoot + "/uninstall.sh), rebuild the hashtab and restart.")
	return nil
}

func uploadTree(c *ssh.Client) error {
	sc, err := sftp.NewClient(c)
	useSFTP := err == nil
	if useSFTP {
		defer sc.Close()
	}

	return fs.WalkDir(payload, "payload", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(p, "payload")
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		dst := path.Join(remoteRoot, rel)
		if d.IsDir() {
			if useSFTP {
				return sc.MkdirAll(dst)
			}
			_, err := runCmd(c, "mkdir -p '"+dst+"'")
			return err
		}
		mode := os.FileMode(0644)
		base := path.Base(rel)
		if base == "install.sh" || base == "uninstall.sh" {
			mode = 0755
		}
		if useSFTP {
			if err := sftpPutFile(sc, p, dst, mode); err != nil {
				return fmt.Errorf("copying %s failed: %v", rel, err)
			}
		} else {
			if err := catPutFile(c, p, dst, mode); err != nil {
				return fmt.Errorf("copying %s failed: %v", rel, err)
			}
		}
		fmt.Println("      copied", rel)
		return nil
	})
}

func sftpPutFile(sc *sftp.Client, src, dst string, mode os.FileMode) error {
	data, err := fs.ReadFile(payload, src)
	if err != nil {
		return err
	}
	if err := sc.MkdirAll(path.Dir(dst)); err != nil {
		return err
	}
	w, err := sc.Create(dst)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := sc.Chmod(dst, mode); err != nil {
		return err
	}
	st, err := sc.Stat(dst)
	if err != nil {
		return err
	}
	if st.Size() != int64(len(data)) {
		return fmt.Errorf("size mismatch (%d of %d bytes)", st.Size(), len(data))
	}
	return nil
}

func catPutFile(c *ssh.Client, src, dst string, mode os.FileMode) error {
	data, err := fs.ReadFile(payload, src)
	if err != nil {
		return err
	}
	if _, err := runCmd(c, "mkdir -p '"+path.Dir(dst)+"'"); err != nil {
		return err
	}
	s, err := c.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	s.Stdin = bytes.NewReader(data)
	var out lockedBuffer
	s.Stdout, s.Stderr = &out, &out
	if err := s.Run(fmt.Sprintf("cat > '%s' && chmod %o '%s' && wc -c < '%s'", dst, mode, dst, dst)); err != nil {
		return fmt.Errorf("%v %s", err, out.String())
	}
	if strings.TrimSpace(out.String()) != fmt.Sprint(len(data)) {
		return fmt.Errorf("size mismatch (%s of %d bytes)", strings.TrimSpace(out.String()), len(data))
	}
	return nil
}

// ---------- uninstall ----------

func uninstall(c *ssh.Client) error {
	fmt.Println()
	if !yesNo("Remove Stickerz from this tablet?", false) {
		fmt.Println("Nothing was changed.")
		return nil
	}
	removeLib := yesNo("Also delete the Stickers library notebook?", false)

	script := remoteRoot + "/uninstall.sh"
	exists, _ := runCmd(c, "if [ -f "+script+" ]; then echo YES; else echo NO; fi")
	if exists != "YES" {
		return errors.New("uninstall.sh was not found at " + script + ".\n" +
			"SSH in and remove /home/root/xovi/exthome/qt-resource-rebuilder/stickerz.qmd\n" +
			"(and rm-stickers.qmd if still present), then rebuild the hashtab and restart in reManager.")
	}
	cmd := "sh " + script
	if removeLib {
		cmd = "sh " + script + " --remove-library"
	}

	fmt.Println()
	fmt.Println("Running uninstall on the tablet...")
	fmt.Println()
	out, err := runCmd(c, cmd)
	if out != "" {
		fmt.Println(out)
		fmt.Println()
	}
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	fmt.Println("Stickerz has been removed (per uninstall.sh).")
	fmt.Println("This installer does NOT restart the UI.")
	fmt.Println("In reManager: rebuild the hashtab, then restart XOVI / xochitl.")
	return nil
}

// lockedBuffer is safe for the concurrent stdout/stderr copiers of an SSH session.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}
