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
	default:
		panic("wat should I do")
	}
}

func parent() {
	fmt.Printf("Running %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
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
