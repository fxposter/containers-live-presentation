# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure('2') do |config|
  config.vm.box = 'ubuntu/trusty64'
  config.vm.synced_folder './code', '/data'

  config.vm.provision 'shell', inline: <<-SHELL
    sudo add-apt-repository ppa:ubuntu-lxc/lxd-stable
    echo 'deb https://apt.dockerproject.org/repo ubuntu-trusty main' > /etc/apt/sources.list.d/docker.list
    sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
    sudo apt-get update -y
    sudo apt-get install -y golang apt-transport-https ca-certificates docker-engine
    # linux-image-extra-$(uname -r) linux-image-extra-virtual
  SHELL


  config.vm.provision 'shell', inline: <<-SHELL
    cd /root
    rm -rf rootfs
    mkdir rootfs
    cd rootfs
    docker create --name alpine-for-export alpine true
    docker export alpine-for-export -o alpine.tar
    docker rm -f alpine-for-export
    tar xvf alpine.tar
    rm alpine.tar
    rmdir proc
    touch I_AM_INSIDE_CONTAINER
  SHELL

  config.vm.provision 'shell', inline: <<-SHELL
    cp /vagrant/setup_container.sh /root/rootfs/
    chroot /root/rootfs /bin/sh /setup_container.sh
  SHELL

  config.vm.provision 'shell', inline: <<-SHELL
    echo '127.0.0.1 virtualbox localhost' > /etc/hosts
    echo 'virtualbox' > /etc/hostname
    hostname -F /etc/hostname
  SHELL
end

# cd ..
# unshare -m /bin/bash
# mount --make-rslave /
# mount --bind `pwd`/rootfs `pwd`/rootfs
# mount --make-private `pwd`/rootfs
# mkdir -p rootfs/oldrootfs
# pivot_root `pwd`/rootfs `pwd`/rootfs/oldrootfs
# cd /
# mkdir -p rootfs/proc
# mount -t proc proc /proc
# mount --make-rprivate /oldrootfs
# umount -l /oldrootfs
# rmdir /oldrootfs
