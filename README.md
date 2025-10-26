# EdgeCD

## Table of Contents

*   [Why EdgeCD?](#why-edgecd)
*   [How does EdgeCD Helps SRE, DevOps, and Systems Engineers](#how-does-edgecd-helps-sre-devops-and-systems-engineers)
    *   [Site Reliability Engineers (SRE)](#site-reliability-engineers-sre)
    *   [DevOps Engineers](#devops-engineers)
    *   [System Engineers](#system-engineers)
*   [Getting Started](#getting-started)
    *   [Installation with `edgectl bootstrap`](#installation-with-edgectl-bootstrap)
    *   [Bootstrap Command Flags](#bootstrap-command-flags)
    *   [Manual Installation](#manual-installation)
*   [Configuration](#configuration)
    *   [`config.yaml` Structure](#configyaml-structure)
    *   [Configuration Options](#configuration-options)

## Why EdgeCD?

The deployment lifecycle, synchronization, and health reconciliation for routers, IoT agents, bare-metal servers or edge computers must be **automated**, **auditable**, and **self-correcting** without relying on push-based updates or manual interventions.

**EdgeCD** automates the entire lifecycle of edge devices by continuously monitoring a Git repository for declarative configuration changes (packages, files, scripts). It performs a full state reconciliation, including dependency checks, file synchronization, service restarts, and conditional device reboots, ensuring that the physical fleet's state always matches the desired state defined in Git.

## How does EdgeCD Helps SRE, DevOps, and Systems Engineers

The functionality of **EdgeCD** provides significant benefits across different engineering disciplines:

### Site Reliability Engineers (SRE)

* **Configuration Drift Prevention:** By continuously checking the device state against the Git repository and enforcing changes via hard-resets and file synchronization, **EdgeCD** virtually **eliminates configuration drift**. This is a core SRE concern, as drift is a major cause of service degradation and outages.
* **Automated Remediation:** If a file is manually modified on the device, **EdgeCD** automatically reverts it to the configured state on the next poll, providing **self-healing capabilities** and reducing the need for human intervention.
* **Reliable Rollouts/Rollbacks:** Since the *entire* state is driven by Git commits, SREs can reliably **roll out** a new configuration by merging to `main` and **roll back** to a previous known-good state simply by resetting the Git branch. **EdgeCD** handles the deployment mechanism on the device.

### DevOps Engineers

* **GitOps Workflow:** **EdgeCD** establishes a **GitOps pipeline** where configuration changes are pulled from the repository. This aligns with modern DevOps principles:
  ***Declarative Configuration:** The router's desired state is declared in a YAML file (`edge-cd.yaml`).
  * **Version Control:** Every change is tracked, reviewed, and audited via Git.
  * **Automated Delivery:** **EdgeCD** automates the continuous pulling, comparison, and application of configuration.
* **Consistency and Scale:** The template-based approach (using the device's hostname to target a specific configuration folder: `clusters/${HOSTNAME}/`) allows DevOps teams to manage a large fleet of heterogeneous devices with a single repository, ensuring **consistency and scalable deployment**.
* **Dependency Management:** The `syncPackages` function automates **package dependency installation** (`opkg`), simplifying the environment setup for new features or security updates.

### System Engineers

* **Standardized Environment:** The `syncStartupScripts` and `syncFiles`/`syncDirectories` functions ensure a **standardized operating environment** across all deployed devices. This makes developing and testing system-level features, like custom boot scripts or configuration templates, more predictable.
* **Atomic Updates:** **EdgeCD** uses commit hashes (`checkCommit`) to determine if a full synchronization is needed. This guarantees that **all related configuration changes (files, services, packages)** within a single Git commit are applied together, preventing partial, unstable deployments.
* **Targeted Actions (Reboot/Service Restart):** Developers can flag specific files to either **restart associated services** or **trigger a full device reboot** upon change. This provides fine-grained control over the deployment impact, which is critical when updating core system files or services.
* **Self-Updating Agent:** The `syncSelf` function ensures the sync agent itself is always the latest version, which simplifies maintenance and feature rollouts for the **infrastructure code (EdgeCD)** itself.

## Getting Started

### Installation with `edgectl bootstrap`

The `edgectl bootstrap` command is the recommended way to install and configure `edge-cd` on a target device. It automates the entire process, including cloning the necessary repositories, installing dependencies, placing the configuration file, and setting up the `edge-cd` service.

**Prerequisites:**

*   An SSH key pair to access the target device.
*   The `edgectl` binary on your local machine.

**Usage:**

```bash
edgectl bootstrap \
  --target-addr <user@host> \
  --ssh-private-key <path/to/your/ssh/key> \
  --config-repo <your-config-repo-url> \
  --packages <package1,package2> \
  --service-manager <systemd|procd> \
  --package-manager <apt|opkg>
```

### Bootstrap Command Flags

The `bootstrap` command accepts the following flags:

| Flag                     | Description                                                                                              | Required |
| ------------------------ | -------------------------------------------------------------------------------------------------------- | -------- |
| `--target-addr`          | The SSH address of the target device (e.g., `user@host`).                                                | Yes      |
| `--ssh-private-key`      | Path to the SSH private key for authentication.                                                          | Yes      |
| `--config-repo`          | The URL of your Git repository containing the `edge-cd` configuration.                                   | Yes      |
| `--target-user`          | The SSH user for the target device (default: `root`).                                                    | No       |
| `--config-path`          | Path to the directory containing the config spec file.                                                   | No       |
| `--config-spec`          | Name of the config spec file.                                                                            | No       |
| `--edge-cd-repo`         | The URL of the `edge-cd` Git repository (default: `https://github.com/alexandremahdhaoui/edge-cd.git`).  | No       |
| `--edgecd-branch`        | The branch name for the `edge-cd` repository (default: `main`).                                          | No       |
| `--config-branch`        | The branch name for the config repository (default: `main`).                                             | No       |
| `--packages`             | A comma-separated list of packages to install on the target device.                                      | No       |
| `--service-manager`      | The service manager to use (`systemd` or `procd`).                                                       | No       |
| `--package-manager`      | The package manager to use (`apt` or `opkg`).                                                            | No       |
| `--edge-cd-repo-dest`    | The destination path for the `edge-cd` repository on the target device.                                  | No       |
| `--user-config-repo-dest`| The destination path for the user config repository on the target device.                                | No       |
| `--inject-prepend-cmd`   | A command to prepend to privileged operations (e.g., `sudo`).                                            | No       |
| `--inject-env`           | Environment variables to inject on the target device (e.g., `GIT_SSH_COMMAND=...`).                      | No       |

### Manual Installation

If you prefer to install `edge-cd` manually, follow these steps:

1.  **SSH into the target device:**

    ```bash
    ssh <user@host>
    ```

2.  **Install dependencies:**

    Install `git` and any other required packages using your device's package manager.

3.  **Clone the repositories:**

    Clone the `edge-cd` repository and your configuration repository:

    ```bash
    git clone https://github.com/alexandremahdhaoui/edge-cd.git /usr/local/src/edge-cd
    git clone <your-config-repo-url> /usr/local/src/edge-cd-config
    ```

4.  **Place the configuration file:**

    Create a `config.yaml` file in `/etc/edge-cd/` with the required configuration (see the "Configuration" section below).

5.  **Set up the service:**

    Copy the appropriate service file from the `edge-cd` repository to your device's service manager directory and enable the service.

    For `systemd`:

    ```bash
    sudo cp /usr/local/src/edge-cd/cmd/edge-cd/service-managers/systemd/service /etc/systemd/system/edge-cd.service
    sudo systemctl daemon-reload
    sudo systemctl enable edge-cd.service
    sudo systemctl start edge-cd.service
    ```

    For `procd`:

    ```bash
    sudo cp /usr/local/src/edge-cd/cmd/edge-cd/service-managers/procd/service /etc/init.d/edge-cd
    /etc/init.d/edge-cd enable
    /etc/init.d/edge-cd start
    ```

## Configuration

`edge-cd` is configured using a `config.yaml` file located at `/etc/edge-cd/config.yaml` on the target device. This file defines the behavior of `edge-cd`, including the repositories to monitor, packages to install, and files to sync.

### `config.yaml` Structure

Here is an example of a `config.yaml` file:

```yaml
# -- defines how EdgeCD clone itself
edgectl:
  autoUpdate:
    enabled: true
  repo:
    url: "https://github.com/alexandremahdhaoui/edge-cd.git"
    branch: "main"
    destinationPath: "/usr/local/src/edge-cd"

config:
  spec: "spec.yaml"
  path: "./devices/${HOSTNAME}"
  repo:
    url: "<your-config-repo-url>"
    branch: "main"
    destinationPath: "/usr/local/src/deployment"

pollingIntervalSecond: 60

extraEnvs:
  - HOME: /root
  - GIT_SSH_COMMAND: "ssh -o StrictHostKeyChecking=accept-new"

serviceManager:
  name: "systemd"

packageManager:
  name: "apt"
  autoUpgrade: false
  requiredPackages:
    - "git"
    - "wget"

# -- Sync directories
directories:
  - source: "/path/to/source"
    destination: "/path/to/destination"
    owner: "user:group"
    permissions: "0644"

# -- Sync single files
files:
  - source: "/path/to/source/file"
    destination: "/path/to/destination/file"
    owner: "user:group"
    permissions: "0644"
```

### Configuration Options

*   `edgectl`: Configures the behavior of the `edgectl` tool.
    *   `autoUpdate`: Enables or disables automatic updates for `edge-cd`.
    *   `repo`: Defines the `edge-cd` repository URL, branch, and destination path.
*   `config`: Defines the user's configuration repository.
    *   `spec`: The name of the configuration spec file.
    *   `path`: The path to the directory containing the device's configuration.
    *   `repo`: Defines the configuration repository URL, branch, and destination path.
*   `pollingIntervalSecond`: The interval in seconds at which `edge-cd` polls the Git repository for changes.
*   `extraEnvs`: A list of environment variables to be set when `edge-cd` runs.
*   `serviceManager`: The name of the service manager to use (`systemd` or `procd`).
*   `packageManager`: The name of the package manager to use (`apt` or `opkg`).
    *   `autoUpgrade`: Enables or disables automatic package upgrades.
    *   `requiredPackages`: A list of packages to be installed.
*   `directories`: A list of directories to sync.
    *   `source`: The source path in the configuration repository.
    *   `destination`: The destination path on the target device.
    *   `owner`: The owner and group of the synced directory.
    *   `permissions`: The permissions of the synced directory.
*   `files`: A list of files to sync.
    *   `source`: The source path in the configuration repository.
    *   `destination`: The destination path on the target device.
    *   `owner`: The owner and group of the synced file.
    *   `permissions`: The permissions of the synced file.

## See Also

*   [Documentation Conventions](./docs/doc-conventions.md)
*   [`edge-cd` Command](./cmd/edge-cd/README.md)
*   [`edgectl` Command](./cmd/edgectl/README.md)
*   [`edgectl-e2e` Command](./cmd/edgectl-e2e/README.md)
