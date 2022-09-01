# Alpine Linux
# https://hub.docker.com/_/alpine
ARG ALPINE_VERSION=3.15.4
FROM alpine:${ALPINE_VERSION}

# Packages
RUN set -eux; \
    apk update; \
    apk add --no-cache alpine-base busybox-extras syslog-ng bash su-exec sudo ca-certificates openssh rsync shadow rssh; \
    mkdir -p /usr/local/sbin;

# OpenRC
RUN set -eux; \
    sed -i -e 's/^\(tty\d\:\:\)/#\1/g' /etc/inittab; \
    sed -i -e 's/#rc_sys=".*"/rc_sys="docker"/g' /etc/rc.conf; \
    sed -i -e 's/#rc_env_allow=".*"/rc_env_allow="\*"/g' /etc/rc.conf; \
    sed -i -e 's/#rc_provide=".*"/rc_provide="loopback net"/g' /etc/rc.conf; \
    rm -f /etc/init.d/hwdrivers /etc/init.d/hwclock /etc/init.d/hwdrivers /etc/init.d/modules /etc/init.d/modules-load /etc/init.d/modloop;

# Syslog
COPY syslog-ng/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf
RUN set -eux; \
    rc-update add syslog-ng default;

# OpenSSH
COPY sshd-config/sshd-config.sh /usr/local/sbin/sshd-config
COPY sshd-config/sshd-config.init /etc/init.d/sshd-config

RUN set -eux; \
    deluser $(getent passwd 33 | cut -d: -f1); \
    delgroup $(getent group 33 | cut -d: -f1) 2>/dev/null || true; \
    mkdir -p /root/.ssh /etc/authorized_keys; \
    chmod 700 /root/.ssh; \
    sed -i -e 's/#command_args=".*"/command_args="-e"/g' /etc/conf.d/sshd; \
    sed -i 's/^#\?\(AuthorizedKeysFile\)\s*.*$/\1 \/etc\/authorized_keys\/%u/' /etc/ssh/sshd_config; \
    sed -i 's/^#\?\(Port\)\s*.*$/\1 22/' /etc/ssh/sshd_config; \
    cp -dR /etc/ssh /etc/ssh.default; \
    chmod 755 /usr/local/sbin/sshd-config; \
    chmod 755 /etc/init.d/sshd-config; \
    rc-update add sshd-config default; \
    rc-update add sshd default; 

EXPOSE 22

# Users
RUN set -eux; \
    addgroup -g 32 -S sudo; \
    echo '%sudo ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/sudo-nopasswd; \ 
    usermod -p 'PUzhTltyGyNEc' root; \
    useradd --uid 1000 -p 'I4tpLGmwuVu2o' someone; \
    useradd --uid 1001 -p '0qnjux6WmcQno' --groups sudo someadmin;

CMD ["/sbin/init"]
