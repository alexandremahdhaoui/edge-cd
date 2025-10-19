# router-sync

`router-sync` can be deployed on a router using [bootstrap-router](../bootstrap-router/README.md).

TODOs:

- `router-sync` REPO_NAME etc... can be specified with the `router-sync.yaml` config file.
- `bootstrap-router` must also install the `router-sync.yaml` config file. (otherwise script does not work)

## Overview 🚀

`router-sync.sh` is a self-updating **configuration synchronization script** designed for routers and embedded systems (likely OpenWrt/LEDE, given the use of `opkg` and `/etc/rc.local`).

It runs continuously, pulling configuration, packages, and scripts from a specified Git repository to maintain the device's state according to the remote configuration. It ensures that configuration files are synchronized only when changes are detected in the repository, and it handles necessary service restarts or device reboots.

---

## Core Features ✨

- **Self-Updating:** The script is capable of updating itself (`router-sync.sh` and the associated init script) from the Git repository.
- **Configuration Management:** Synchronizes individual configuration **files** and entire **directories** from the Git repository to local paths (e.g., `/etc/config`).
- **Package Management:** Installs required packages defined in the configuration using **`opkg`**.
- **Startup Script Sync:** Updates the contents of the local `/etc/rc.local` file with custom startup commands.
- **Idempotency:** Utilizes commit hashes to skip the synchronization process if the configuration has not changed since the last successful run.
- **Conditional Actions:** Triggers service restarts or a full device reboot if file changes require it.

---

## Dependencies 📦

This script requires the following tools to be installed on the router:

- **`bash`**: The shell environment.
- **`git`**: For cloning, fetching, and checking out the configuration repository.
- **`yq`**: A lightweight YAML processor used to read and parse the configuration file (`router-sync.yaml`).
- **`opkg`**: The package manager used to install required software.
- **`ssh`**: Required for secure access to the private Git repository.

---

## Configuration ⚙️

The script reads its configuration from a YAML file specific to the router's hostname, located within the cloned repository:

- **Repository Path**: `/tmp/deployment` (temporary path for the Git clone).
- **Configuration File**: `/tmp/deployment/tiers/bootstrap/$(HOSTNAME)/router-sync.yaml`

The configuration file controls:

- `pollingIntervalSecond`: How long the script sleeps between sync attempts.
- `requiredPackages`: A list of packages to install via `opkg`.
- `startupScripts`: A list of commands to be written to `/etc/rc.local`.
- `files`: A list of individual files to sync, including their source, destination, permissions (`fileMod`), services to restart (`restartServices`), and whether a reboot is required (`rebootOnChange`).
- `directories`: A list of directories to sync, including mappings for service restarts based on the individual file name.

### Configuration Example

This example snippet from the YAML shows:

1. The sync interval is **60 seconds**.
2. The **`quagga-bgpd`** and **`quagga-zebra`** packages will be installed.
3. Changes to `config/network` will trigger the **`network`** service to restart.
4. Changes to `config/system` will trigger a **full device reboot** (`rebootOnChange`).

```yaml
pollingIntervalSecond: 60

requiredPackages:
  - quagga-bgpd
  - quagga-zebra

startupScripts: []

# -- Sync directories
directories:
  # -- Binaries
  - sourceDir: bin
    destDir: /usr/bin
    fileMod: 755
    restartServicesMap: {}

  # -- OpenWRT Config
  - sourceDir: config
    destDir: /etc/config
    fileMod: 644
    restartServicesMap:
      dhcp:
        - odhcpd
        - dnsmasq
      firewall:
        - firewall
      network:
        - network
      router-sync:
        - router-sync
    rebootOnChange:
      - system

  # -- Services
  - sourceDir: services
    destDir: /etc/init.d
    fileMod: 755
    restartServicesMap: {}

# -- Sync Files
files:
  - sourcePath: files/quagga_bgpd.conf
    destPath: /etc/quagga/bgpd.conf
    restartServices:
      - quagga
  - sourcePath: files/quagga_zebra.conf
    destPath: /etc/quagga/zebra.conf
    restartServices:
      - quagga
```

---

## Workflow Summary 🔄

1. **Initialization**: Sets environment variables and configuration paths.
2. **Repository Sync**: Clones the remote Git repository (or performs a hard reset if already cloned).
3. **Self-Check**: Updates the `router-sync.sh` script itself and checks if a new script version was synchronized. If so, it exits cleanly for the init system to restart the new version.
4. **Commit Check**: Compares the current commit hash with the last synchronized commit hash to prevent re-synchronizing unchanged configuration.
5. **Configuration Tasks**: Executes package installation, startup script update, and file/directory synchronization based on `router-sync.yaml`.
6. **Action**: If file changes required a reboot, the device reboots. Otherwise, all marked services are restarted.
7. **Sleep**: Pauses for the duration specified by `pollingIntervalSecond` before starting the loop again.

---

## Debug router-sync

On the router run:

```bash
logread | grep "router-sync"
```
