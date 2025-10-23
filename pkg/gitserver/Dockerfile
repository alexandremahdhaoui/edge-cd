FROM alpine:latest

RUN apk update && \
    apk add --no-cache openssh git bash

RUN adduser -D git -s /usr/bin/git-shell && \
    mkdir -p /home/git/.ssh && \
    chmod 0700 /home/git/.ssh && \
    chown -R git:git /home/git

RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin no/g' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/g' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/g' /etc/ssh/sshd_config && \
    echo "AllowUsers git" >> /etc/ssh/sshd_config

EXPOSE 22

COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]
