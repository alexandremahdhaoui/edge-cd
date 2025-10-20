# -- https://hub.docker.com/r/openwrt/rootfs/tags
FROM openwrt/rootfs:x86_64-v21.02.7

ENV PATH="/usr/sbin:/usr/bin:/sbin:/bin"

RUN mkdir -p /var/lock && \
    opkg update && \
    opkg install git git-http procps-ng-ps bash wget openssh-client-utils coreutils-timeout && \
    wget https://github.com/mikefarah/yq/releases/download/v4.30.8/yq_linux_amd64 -O /usr/bin/yq && \
    chmod +x /usr/bin/yq && \
    rm -rf /var/opkg-lists/*

RUN git config --global user.email "e2e-test@edge-cd.com" && \
    git config --global user.name "e2e-test" && \
    git config --global --add safe.directory /opt/config

RUN mkdir -p /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts

COPY test/edge-cd/e2e/test-runner.sh /usr/local/bin/test-runner.sh
COPY test/edge-cd/e2e/config /opt/config
COPY . /opt/src/edge-cd
RUN chmod +x /opt/src/edge-cd/cmd/edge-cd/edge-cd
RUN chmod +x /usr/local/bin/test-runner.sh

WORKDIR /opt

CMD [ "/usr/local/bin/test-runner.sh" ]