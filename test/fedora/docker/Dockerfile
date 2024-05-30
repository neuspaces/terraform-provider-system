# Fedora
# https://hub.docker.com/_/fedora
ARG FEDORA_VERSION=40
FROM fedora:${FEDORA_VERSION}

# Packages
RUN set -eux; \
    dnf install -y systemd openssh-server rsync passwd busybox sudo which; \ 
    dnf clean all; 

# Systemd
RUN set -eux; \
    systemctl set-default multi-user.target; \
    systemctl disable getty@tty1.service;

# OpenSSH
COPY sshd-config/sshd-config.sh /usr/local/sbin/sshd-config
COPY sshd-config/sshd-config.service /etc/systemd/system/sshd-config.service

RUN set -eux; \
    mkdir -p /root/.ssh /etc/authorized_keys; \
    chmod 700 /root/.ssh; \
    sed -i 's/^#\?\(AuthorizedKeysFile\)\s*.*$/\1 .ssh\/authorized_keys \/etc\/authorized_keys\/%u/' /etc/ssh/sshd_config; \
    sed -i 's/^#\?\(Port\)\s*.*$/\1 22/' /etc/ssh/sshd_config; \
    cp -dR /etc/ssh /etc/ssh.default; \
    chmod 755 /usr/local/sbin/sshd-config; \
    systemctl enable sshd-config.service;

EXPOSE 22

# Users
RUN set -eux; \
    groupadd --system www-data; \
    useradd --system --gid www-data --shell /usr/sbin/nologin --home-dir /var/www --no-create-home www-data; \
    echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/sudo-nopasswd; \ 
    usermod -p 'PUzhTltyGyNEc' root; \
    useradd --uid 1000 -p 'I4tpLGmwuVu2o' someone; \
    useradd --uid 1001 -p '0qnjux6WmcQno' --groups wheel someadmin;

# Configure systemd for running within the context of a Docker container
ENV container=docker

# Configure signal provided to systemd when the container is stopped 
STOPSIGNAL SIGRTMIN+3

# systemd requires access to the cgroup hierarchies at `/sys/fs/cgroup`
# `/run`, `/run/lock`, and `/tmp` are required to be tmpfs
VOLUME [ "/sys/fs/cgroup", "/run", "/run/lock", "/tmp" ]

CMD ["/sbin/init"]
