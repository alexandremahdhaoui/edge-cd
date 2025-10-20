# e2e test

### Test grid

| OS      | Package Manager | Service Manage | Arch |
|---------|-----------------|----------------|------|
| openwrt | opkg            | procd          | x86  |

### Docs

-[github.com/openwrt/docker](https://github.com/openwrt/docker)

```bash
# -- https://hub.docker.com/r/openwrt/rootfs/tags
docker run -i --rm openwrt/rootfs:x86_64-v21.02.7 <<'EOF'
[ ! -d ./scripts ] && ./setup.sh
mkdir /var/lock/
opkg update
EOF
```
