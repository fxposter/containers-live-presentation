# Step 0

* docker pull busybox

* docker run --rm -it busybox /bin/sh
  * ls -al /
  * ps ax
  * hostname

# Step 1 - exec.Command

```
cmd := exec.Command(os.Args[2], os.Args[3:]...)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
must(cmd.Run())
```

* go run scratch.go run echo true

* hostname
* go run scratch.go run /bin/bash
  * hostname
  * hostname example.com
  * hostname
* hostname
* hostname virtualbox

# Step 2 - syscall.SysProcAttr

```
cmd.SysProcAttr = &syscall.SysProcAttr{
  Cloneflags: syscall.CLONE_NEWUTS,
}
```

* hostname
* go run scratch.go run /bin/bash
  * hostname
  * hostname example.com
  * hostname
* hostname

# Step 3 - show PID

```
fmt.Printf("Running %v on PID %d\n", os.Args[2:], os.Getpid())
```

* go run scratch.go run /bin/bash
  * ps ax | grep PID

# Step 4 - syscall.CLONE_NEWPID

```
cmd.SysProcAttr = &syscall.SysProcAttr{
  Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
}
```

* go run scratch.go run /bin/bash
  * ps ax | grep PID

# Step 5 - child()

```
case "child":
  child()
```

* go run scratch.go run echo hello
* go run scratch.go run /bin/bash
  * ps ax
* ls -al /proc


# Step 6 - syscall.Chroot/os.Chdir

```
must(syscall.Chroot("/root/rootfs"))
must(os.Chdir("/"))
```

* go run scratch.go run /bin/bash
  * ls -al
  * ps ax
  * mount

# Step 7 - hostname

```
shellOut("hostname -F /etc/hostname")
```

* go run scratch.go run /bin/bash
  * hostname

# Step 8 - syscall.Mount("", "/proc", "proc", 0, "")

```
must(os.MkdirAll("/proc", 0555))
must(syscall.Mount("", "/proc", "proc", 0, ""))
```

* go run scratch.go run /bin/bash
  * ls -al
  * ps ax
  * mount

# Step 9 - unchroot

* gcc -static -o /root/rootfs/unchroot unchroot.c
* go run scratch.go run /bin/bash
  * ls -al
  * ./unchroot
    * ls -al

# Step 10

* cat /proc/self/mountinfo
* umount /root/rootfs/proc (many times)

# Step 11 (do not try this at home) - syscall.PivotRoot

```
must(syscall.Mount("/root/rootfs", "/root/rootfs", "", syscall.MS_BIND, ""))
must(os.MkdirAll("/root/rootfs/oldrootfs", 0700))
must(syscall.PivotRoot("/root/rootfs", "/root/rootfs/oldrootfs"))
```

* go run scratch.go run /bin/bash
  * ls -al
  * ls -al oldrootfs
  * mount
  * ./unchroot
* cat /proc/self/mountinfo
* cd /
* ls -al

vagrant reload

# Step 12 - syscall.CLONE_NEWNS

```
cmd.SysProcAttr = &syscall.SysProcAttr{
  Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
}
```

* go run scratch.go run /bin/bash
  * ls -al
  * ls -al oldrootfs
  * mount
  * ./unchroot
* cat /proc/self/mountinfo

# Step 13 - syscall.Unmount /oldrootfs

```
must(syscall.Mount("", "/oldrootfs", "", syscall.MS_PRIVATE|syscall.MS_REC, ""))
must(syscall.Unmount("/oldrootfs", syscall.MNT_DETACH))
must(os.Remove("/oldrootfs"))
```

* go run scratch.go run /bin/bash
  * ls -al
  * mount
* mount

# Step 14 - check for root process

```
import "os/signal"

go func() {
  c := make(chan os.Signal, 1)
  signal.Notify(c, syscall.SIGTERM)
  s := <-c
  fmt.Println("Got signal:", s)
}()
```

* go run scratch.go run /bin/bash
  * ps ax
* go run scratch.go run /bin/sleep 1000
* kill {PID}
* go run scratch.go run ps ax

# Step 15 - syscall.Exec

```
binary, err := exec.LookPath(os.Args[2])
must(err)
must(syscall.Exec(binary, os.Args[2:], os.Environ()))
```

* go run scratch.go run ps ax
* go run scratch.go run /bin/bash
  * ps ax

# Step 16 - veth

```
must(cmd.Start())

shellOut("ip link add veth0 type veth peer name veth1")
shellOut("ifconfig veth0 10.0.0.1/24 up")
shellOut("ln -s /proc/" + strconv.Itoa(cmd.Process.Pid) + "/ns/net /var/run/netns/container")
shellOut("ip link set veth1 netns container")
shellOut("ip netns exec container ifconfig lo up")
shellOut("ip netns exec container ifconfig veth1 10.0.0.2/24 up")

err := cmd.Wait()
must(os.RemoveAll("/var/run/netns/container"))
must(err)
```

* go run scratch.go run /bin/bash
  * ping 10.0.0.1
* ping 10.0.0.2
* go run scratch.go run /bin/bash
  * echo "echo 'Hello'" > /echo
  * chmod a+x /echo
  * while true ; do nc -lp 80 -e /echo ; done
* telnet 10.0.0.2 80
  * ping 8.8.8.8

# Step 17

```
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
```

* go run scratch.go run /bin/bash
  * ping 8.8.8.8
  * ping ya.ru

# Step 18

```
must(ioutil.WriteFile("/etc/resolv.conf", []byte("nameserver 8.8.8.8"), 0777))
```

* go run scratch.go run /bin/bash
  * ping ya.ru


* https://lwn.net/Articles/252794/
* https://lwn.net/Articles/689856/
* https://lwn.net/Articles/531114/
* https://lwn.net/Articles/580893/
* http://www.ibm.com/developerworks/linux/library/l-mount-namespaces/index.html
