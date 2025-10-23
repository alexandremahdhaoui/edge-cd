#!/bin/bash

ssh-keygen -A

if [ ! -d "/srv/git" ]; then
    mkdir -p /srv/git
fi

if [ -f /tmp/authorized_keys ]; then
    cp /tmp/authorized_keys /home/git/.ssh/
    chmod 600 /home/git/.ssh/authorized_keys
    chown git:git /home/git/.ssh/authorized_keys
    rm -f /tmp/authorized_keys
fi

exec /usr/sbin/sshd -D -e
