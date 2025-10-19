# EdgeCD

## Why EdgeCD?

Configuration files, device packages, and environment variables for embedded systems must be **declarative** and **version controlled**.

The deployment lifecycle, synchronization, and health reconciliation for distributed routers and IoT agents must be **automated**, **auditable**, and **self-correcting** without relying on push-based updates or manual interventions.

**EdgeCD** automates the entire lifecycle of edge devices by continuously monitoring a Git repository for declarative configuration changes (packages, files, scripts). It performs a full state reconciliation, including dependency checks, file synchronization, service restarts, and conditional device reboots, ensuring that the physical fleet's state always matches the desired state defined in Git.

## How does EdgeCD Helps SRE, DevOps, and Systems Engineers

The functionality of **EdgeCD** provides significant benefits across different engineering disciplines:

### Site Reliability Engineers (SRE)

* **Configuration Drift Prevention:** By continuously checking the device state against the Git repository and enforcing changes via hard-resets and file synchronization, **EdgeCD** virtually **eliminates configuration drift**. This is a core SRE concern, as drift is a major cause of service degradation and outages.
* **Automated Remediation:** If a file is manually modified on the device, **EdgeCD** automatically reverts it to the configured state on the next poll, providing **self-healing capabilities** and reducing the need for human intervention.
* **Reliable Rollouts/Rollbacks:** Since the *entire* state is driven by Git commits, SREs can reliably **roll out** a new configuration by merging to `main` and **roll back** to a previous known-good state simply by resetting the Git branch. **EdgeCD** handles the deployment mechanism on the device.

---

### DevOps Engineers

* **GitOps Workflow:** **EdgeCD** establishes a **GitOps pipeline** where configuration changes are pulled from the repository. This aligns with modern DevOps principles:
  ***Declarative Configuration:** The router's desired state is declared in a YAML file (`router-sync.yaml`).
  * **Version Control:** Every change is tracked, reviewed, and audited via Git.
  * **Automated Delivery:** **EdgeCD** automates the continuous pulling, comparison, and application of configuration.
* **Consistency and Scale:** The template-based approach (using the device's hostname to target a specific configuration folder: `clusters/${HOSTNAME}/`) allows DevOps teams to manage a large fleet of heterogeneous devices with a single repository, ensuring **consistency and scalable deployment**.
* **Dependency Management:** The `syncPackages` function automates **package dependency installation** (`opkg`), simplifying the environment setup for new features or security updates.

---

### System Engineers

* **Standardized Environment:** The `syncStartupScripts` and `syncFiles`/`syncDirectories` functions ensure a **standardized operating environment** across all deployed devices. This makes developing and testing system-level features, like custom boot scripts or configuration templates, more predictable.
* **Atomic Updates:** **EdgeCD** uses commit hashes (`checkCommit`) to determine if a full synchronization is needed. This guarantees that **all related configuration changes (files, services, packages)** within a single Git commit are applied together, preventing partial, unstable deployments.
* **Targeted Actions (Reboot/Service Restart):** Developers can flag specific files to either **restart associated services** or **trigger a full device reboot** upon change. This provides fine-grained control over the deployment impact, which is critical when updating core system files or services.
* **Self-Updating Agent:** The `syncSelf` function ensures the sync agent itself is always the latest version, which simplifies maintenance and feature rollouts for the **infrastructure code (EdgeCD)** itself.
