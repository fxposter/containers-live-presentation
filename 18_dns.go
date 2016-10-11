package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
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
	must(os.MkdirAll("/var/run/netns", 0755))

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Start())

	shellOut("ip link add veth0 type veth peer name veth1")
	shellOut("ifconfig veth0 10.0.0.1/24 up")
	shellOut("ln -s /proc/" + strconv.Itoa(cmd.Process.Pid) + "/ns/net /var/run/netns/container")
	shellOut("ip link set veth1 netns container")
	shellOut("ip netns exec container ifconfig lo up")
	shellOut("ip netns exec container ifconfig veth1 10.0.0.2/24 up")
	shellOut("ip netns exec container ip route add default via 10.0.0.1")
	shellOut("iptables -t nat -A POSTROUTING -s 10.0.0.0/255.255.255.0 -o eth0 -j MASQUERADE")
	shellOut("iptables -A FORWARD -i eth0 -o veth0 -j ACCEPT")
	shellOut("iptables -A FORWARD -o eth0 -i veth0 -j ACCEPT")

	err := cmd.Wait()
	must(os.RemoveAll("/var/run/netns/container"))
	shellOut("iptables -t nat -D POSTROUTING -s 10.0.0.0/255.255.255.0 -o eth0 -j MASQUERADE")
	shellOut("iptables -D FORWARD -i eth0 -o veth0 -j ACCEPT")
	shellOut("iptables -D FORWARD -o eth0 -i veth0 -j ACCEPT")
	must(err)
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
	must(ioutil.WriteFile("/etc/resolv.conf", []byte("nameserver 8.8.8.8"), 0777))

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
