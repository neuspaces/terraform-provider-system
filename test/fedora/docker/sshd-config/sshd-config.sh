#!/usr/bin/env bash

# The MIT License (MIT)

# Copyright (c) 2015-2021 Volt Grid Pty Ltd
# Copyright (c) 2022 Dominik Lekse

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

set -e

# Empty line
echo ""

# Copy default config from cache, if required
if [ ! "$(ls -A /etc/ssh)" ]; then
    cp -dR /etc/ssh.default/* /etc/ssh/
fi

set_hostkeys() {
    sed -i -e 's/^#\?\(HostKey\)\s*.*_rsa_key$/\1 \/etc\/ssh\/keys\/ssh_host_rsa_key/' /etc/ssh/sshd_config
    sed -i -e 's/^#\?\(HostKey\)\s*.*_ecdsa_key$/\1 \/etc\/ssh\/keys\/ssh_host_ecdsa_key/' /etc/ssh/sshd_config
    sed -i -e 's/^#\?\(HostKey\)\s*.*_ed25519_key$/\1 \/etc\/ssh\/keys\/ssh_host_ed25519_key/' /etc/ssh/sshd_config
}

print_fingerprints() {
    local BASE_DIR=${1-'/etc/ssh'}
    for item in rsa ecdsa ed25519; do
        echo ">>> Fingerprints for ${item} host key"
        ssh-keygen -E md5 -lf ${BASE_DIR}/ssh_host_${item}_key
        ssh-keygen -E sha256 -lf ${BASE_DIR}/ssh_host_${item}_key
        ssh-keygen -E sha512 -lf ${BASE_DIR}/ssh_host_${item}_key
    done
}

check_authorized_key_ownership() {
    local file="$1"
    local _uid="$2"
    local _gid="$3"
    local uid_found="$(stat -c %u ${file})"
    local gid_found="$(stat -c %g ${file})"

    if ! ( [[ ( "$uid_found" == "$_uid" ) && ( "$gid_found" == "$_gid" ) ]] || [[ ( "$uid_found" == "0" ) && ( "$gid_found" == "0" ) ]] ); then
        echo "WARNING: Incorrect ownership for ${file}. Expected uid/gid: ${_uid}/${_gid}, found uid/gid: ${uid_found}/${gid_found}. File uid/gid must match SSH_USERS or be root owned."
    fi
}

# Generate Host keys, if required
if ls /etc/ssh/keys/ssh_host_* 1> /dev/null 2>&1; then
    echo ">> Found host keys in keys directory"
    set_hostkeys
    print_fingerprints /etc/ssh/keys
elif ls /etc/ssh/ssh_host_* 1> /dev/null 2>&1; then
    echo ">> Found Host keys in default location"
    # Don't do anything
    print_fingerprints
else
    echo ">> Generating new host keys"
    mkdir -p /etc/ssh/keys
    ssh-keygen -A
    mv /etc/ssh/ssh_host_* /etc/ssh/keys/
    set_hostkeys
    print_fingerprints /etc/ssh/keys
fi

# Fix permissions, if writable.
# NB ownership of /etc/authorized_keys are not changed
if [ -w /root/.ssh ]; then
    chown root:root /root/.ssh && chmod 700 /root/.ssh/
fi
if [ -w /root/.ssh/authorized_keys ]; then
    chown root:root /root/.ssh/authorized_keys
    chmod 600 /root/.ssh/authorized_keys
fi
if [ -w /etc/authorized_keys ]; then
    chown root:root /etc/authorized_keys
    chmod 755 /etc/authorized_keys
    # test for writability before attempting chmod
    for f in $(find /etc/authorized_keys/ -type f -maxdepth 1); do
        [ -w "${f}" ] && chmod 600 "${f}"
        # Warn if file mode is not 600
        fm=$(stat -c '%a' "${f}")
        if [ "${fm}" != "600" ]; then
            echo "WARNING: ${f} must have permissions 600, but has ${fm}!"
        fi
    done
fi

# Add users if SSH_USERS=user:uid:gid set
if [ -n "${SSH_USERS}" ]; then
    USERS=$(echo $SSH_USERS | tr "," "\n")
    for U in $USERS; do
        IFS=':' read -ra UA <<< "$U"
        _NAME=${UA[0]}
        _UID=${UA[1]}
        _GID=${UA[2]}
        if [ ${#UA[*]} -ge 4 ]; then
            _SHELL=${UA[3]}
        else
            _SHELL=''
        fi

        echo ">> Adding user ${_NAME} with uid: ${_UID}, gid: ${_GID}, shell: ${_SHELL:-<default>}."
        if [ ! -e "/etc/authorized_keys/${_NAME}" ]; then
            echo "WARNING: No SSH authorized_keys found for ${_NAME}!"
        else
            check_authorized_key_ownership /etc/authorized_keys/${_NAME} ${_UID} ${_GID}
        fi
        getent group ${_NAME} >/dev/null 2>&1 || groupadd -g ${_GID} ${_NAME}
        getent passwd ${_NAME} >/dev/null 2>&1 || useradd -r -m -p '' -u ${_UID} -g ${_GID} -s ${_SHELL:-""} -c 'SSHD User' ${_NAME}
    done
else
    # Warn if no authorized_keys
    if [ ! -e /root/.ssh/authorized_keys ] && [ ! "$(ls -A /etc/authorized_keys)" ]; then
        echo "WARNING: No SSH authorized_keys found!"
    fi
fi

# Update MOTD
if [ -v MOTD ]; then
    echo -e "$MOTD" > /etc/motd
fi

# PasswordAuthentication (disabled by default)
if [[ "${SSH_ENABLE_PASSWORD_AUTH}" == "true" ]]; then
    sed -i -e 's/^#\?\(PasswordAuthentication\)\s*.*$/\1 yes/' /etc/ssh/sshd_config
    echo "WARNING: password authentication enabled."

    # Root Password Authentification
    if [[ "${SSH_ENABLE_ROOT_PASSWORD_AUTH}" == "true" ]]; then
        sed -i -e 's/^#\?\(PermitRootLogin\)\s*.*$/\1 yes/' /etc/ssh/sshd_config
        echo "WARNING: password authentication for root user enabled."
    else
        echo "INFO: password authentication is not enabled for the root user. Set SSH_ENABLE_ROOT_PASSWORD_AUTH=true to enable."
    fi

else
    sed -i -e 's/^#\?\(PasswordAuthentication\)\s*.*$/\1 no/' /etc/ssh/sshd_config
    echo "INFO: password authentication is disabled by default. Set SSH_ENABLE_PASSWORD_AUTH=true to enable."
fi

configure_sftp_only_mode() {
    echo "INFO: configuring sftp only mode"
    : ${SFTP_CHROOT:='/data'}
    chown 0:0 ${SFTP_CHROOT}
    chmod 755 ${SFTP_CHROOT}
    sed -i -e 's/^#\?\(Subsystem\s*sftp\)\s*.*$/\1 internal-sftp/' /etc/ssh/sshd_config
    sed -i -e 's/^#\?\(AllowTCPForwarding\)\s*.*$/\1 no/' /etc/ssh/sshd_config
    sed -i -e 's/^#\?\(GatewayPorts\)\s*.*$/\1 no/' /etc/ssh/sshd_config
    sed -i -e 's/^#\?\(X11Forwarding\)\s*.*$/\1 no/' /etc/ssh/sshd_config
    sed -i -e 's/^#\?\(ChrootDirectory\)\s*.*$/\1 '"${SFTP_CHROOT//\//\\\/}"'/' /etc/ssh/sshd_config
}

configure_scp_only_mode() {
    echo "INFO: configuring scp only mode"
    USERS=$(echo $SSH_USERS | tr "," "\n")
    for U in $USERS; do
        _NAME=$(echo "${U}" | cut -d: -f1)
        usermod -s '/usr/bin/rssh' ${_NAME}
    done
    (grep '^[a-zA-Z]' /etc/rssh.conf.default; echo "allowscp") > /etc/rssh.conf
}

configure_rsync_only_mode() {
    echo "INFO: configuring rsync only mode"
    USERS=$(echo $SSH_USERS | tr "," "\n")
    for U in $USERS; do
        _NAME=$(echo "${U}" | cut -d: -f1)
        usermod -s '/usr/bin/rssh' ${_NAME}
    done
    (grep '^[a-zA-Z]' /etc/rssh.conf.default; echo "allowrsync") > /etc/rssh.conf
}

configure_ssh_options() {
    # Enable AllowTcpForwarding
    if [[ "${TCP_FORWARDING}" == "true" ]]; then
        sed -i -e 's/^#\?\(AllowTcpForwarding\)\s*.*$/\1 yes/' /etc/ssh/sshd_config
    fi
    # Enable GatewayPorts
    if [[ "${GATEWAY_PORTS}" == "true" ]]; then
        sed -i -e 's/^#\?\(GatewayPorts\)\s*.*$/\1 yes/' /etc/ssh/sshd_config
    fi
    # Disable SFTP
    if [[ "${DISABLE_SFTP}" == "true" ]]; then
        sed -i -e 's/^#\?\(Subsystem\s*sftp.*\)$/#\1/' /etc/ssh/sshd_config
    fi
}

# Configure mutually exclusive modes
if [[ "${SFTP_MODE}" == "true" ]]; then
    configure_sftp_only_mode
elif [[ "${SCP_MODE}" == "true" ]]; then
    configure_scp_only_mode
elif [[ "${RSYNC_MODE}" == "true" ]]; then
    configure_rsync_only_mode
else
    configure_ssh_options
fi
