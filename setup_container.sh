echo 'nameserver 8.8.8.8' > /etc/resolv.conf
apk update
apk add -u bash
echo '' > /etc/resolv.conf
echo '127.0.0.1 demo-container localhost' > /etc/hosts
echo 'demo-container' > /etc/hostname
hostname -F /etc/hostname
echo 'export PS1="\[\e[31m\]\u\[\e[m\]@\[\e[32m\]\h\[\e[m\]\[\e[34m\]:\w\[\e[m\] \\$ "' > ~/.bashrc
rm /setup_container.sh
