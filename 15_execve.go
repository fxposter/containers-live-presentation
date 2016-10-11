package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 3 {
		panic("use `run` command [arg1 arg2 ...]")
	}

	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("wat should I do")
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v on PID %d\n", os.Args[2:], os.Getpid())

	must(syscall.Mount("/root/rootfs", "/root/rootfs", "", syscall.MS_BIND, ""))
	must(os.MkdirAll("/root/rootfs/oldrootfs", 0700))
	must(syscall.PivotRoot("/root/rootfs", "/root/rootfs/oldrootfs"))

	must(os.Chdir("/"))
	must(os.MkdirAll("/proc", 0555))
	must(syscall.Mount("", "/proc", "proc", 0, ""))

	must(syscall.Mount("", "/oldrootfs", "", syscall.MS_PRIVATE|syscall.MS_REC, ""))
	must(syscall.Unmount("/oldrootfs", syscall.MNT_DETACH))
	must(os.Remove("/oldrootfs"))

	shellOut("hostname -F /etc/hostname")

	binary, err := exec.LookPath(os.Args[2])
	must(err)

	must(syscall.Exec(binary, os.Args[2:], os.Environ()))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func shellOut(command string) {
	parts := strings.Split(command, " ")
	err := exec.Command(parts[0], parts[1:]...).Run()
	if err != nil {
		fmt.Println(err)
	}
}
