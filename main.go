package main
import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"log"
	"os"
	"syscall"
	"unsafe"
	"os/exec"
	"github.com/creack/pty"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func handler(s ssh.Session) {
	io.WriteString(s, "You are connected over ssh\n")
	ptyReq, winCh, isPty := s.Pty()
	if isPty {
		io.WriteString(s, "We Got A terminal babe\n")
		cmd := exec.Command("docker", "exec", "-it", "great_hellman", "/bin/bash")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		io.WriteString(s, fmt.Sprintf("ENV\n%v\n\n", cmd.Env))
		f, err := pty.Start(cmd)
		if err != nil {
			panic(err)
		}
		go func() {
			for win := range winCh {
				setWinsize(f, win.Width, win.Height)
			}
		}()
		go func() {
			io.Copy(f, s) // stdin
		}()
		io.Copy(s, f) // stdout
		cmd.Wait()
	} else {
		io.WriteString(s, "This is bullshit :'(\n")
	}
}
func main() {
	fmt.Println("Welcome to CODEX!")
	ssh.Handle(handler)
	log.Fatal(ssh.ListenAndServe("0.0.0.0:2222", nil))
}
