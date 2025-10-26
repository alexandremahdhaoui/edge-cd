FROM debian:stable-slim

# Install necessary packages: openssh-server for SSH access
# and sudo for potential privilege escalation in tests if needed.
RUN apt-get update && apt-get install -y openssh-server sudo && rm -rf /var/lib/apt/lists/*

# Allow root login with password (for testing purposes)
RUN echo 'PermitRootLogin yes' >> /etc/ssh/sshd_config
RUN echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config
# Accept environment variables needed for git operations
RUN echo 'AcceptEnv GIT_SSH_COMMAND' >> /etc/ssh/sshd_config

# Set a default password for root (for testing purposes)
RUN echo 'root:root' | chpasswd

# Create .ssh directory for root and set permissions
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh

# Add SSH public key from build argument to authorized_keys
ARG SSH_PUBLIC_KEY
RUN echo "${SSH_PUBLIC_KEY}" >> /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys

EXPOSE 22



CMD ["/usr/sbin/sshd", "-D"]

# NOTE: This Dockerfile does not include 'opkg' as it's not available in Debian's default repositories.
# For tasks requiring 'opkg', a more specialized OpenWrt-based Docker image or a mock 'opkg'
# implementation will be necessary. For the initial E2E framework setup, a basic SSH target is sufficient.