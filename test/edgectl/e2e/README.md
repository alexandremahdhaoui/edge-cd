# E2E Testing Guide

## Overview

The e2e tests verify that the complete edge-cd bootstrap pipeline works end-to-end. Tests create real VMs via libvirt, setup a git server, and execute the bootstrap command.

## Quick Start

### Default Test (Automatic Cleanup)

Recommended for CI/CD and automated testing:

```bash
go test ./test/edgectl/e2e -v
```

- Creates test environment with unique ID
- Spins up target VM and git server VM
- Runs bootstrap command
- Destroys VMs and cleans up all artifacts
- Removes test data from system
- Exit code 0 = success, non-zero = failure

### Keep Artifacts for Debugging

When tests fail and you need to inspect the VMs:

```bash
go test ./test/edgectl/e2e -v -args --keep-artifacts
```

- Creates test environment
- Runs full test suite
- **Keeps VMs running** after test completes
- **Keeps artifact directory** with SSH keys and logs
- Saves environment metadata to `~/.edge-cd/e2e/artifacts.json`
- Prints artifact information to test output

### Using the CLI Tool

For more control over the test environment lifecycle:

```bash
# Setup: Create and provision test environment
edgectl-e2e setup
# Output: e2e-20231025-abc123 (save this ID)

# Inspect VMs while they're running
ssh -i /tmp/edge-cd-e2e-abc123/id_rsa_host ubuntu@192.168.1.100

# Run tests in the environment
edgectl-e2e run e2e-20231025-abc123

# Cleanup when done
edgectl-e2e teardown e2e-20231025-abc123
```

### One-Shot CLI Test

Equivalent to default test run:

```bash
edgectl-e2e test
```

## Artifact Structure

When artifacts are retained, you'll find:

```
~/.edge-cd/e2e/artifacts.json      # Metadata for all test environments

/tmp/edge-cd-e2e-<TEST_ID>/        # Per-test artifacts
├── id_rsa_host                     # Private key: edgectl → target VM
├── id_rsa_host.pub
├── id_rsa_target                   # Private key: target VM → git server
├── id_rsa_target.pub
├── setup.log                       # Log of VM/git server setup
├── test.log                        # Log of bootstrap test execution
└── git-server/                     # Git server artifacts
    ├── repos/
    └── ssh/
```

## Manual VM Access

### SSH into Target VM

```bash
# Find the test ID and artifact path
export TEST_ID=e2e-20231025-abc123
export ARTIFACT_PATH=~/.edge-cd/e2e/artifacts.json

# Get VM IP from artifact info
export VM_IP=$(cat ~/.edge-cd/e2e/artifacts.json | jq -r ".environments[] | select(.id == \"$TEST_ID\") | .target_vm.ip")

# SSH in
ssh -i /tmp/edge-cd-e2e-$TEST_ID/id_rsa_host ubuntu@$VM_IP
```

### View Logs

```bash
# Setup phase logs
tail -f /tmp/edge-cd-e2e-$TEST_ID/setup.log

# Test execution logs
tail -f /tmp/edge-cd-e2e-$TEST_ID/test.log
```

### Clone from Git Server

```bash
# Get git server IP
export GIT_IP=$(cat ~/.edge-cd/e2e/artifacts.json | jq -r ".environments[] | select(.id == \"$TEST_ID\") | .git_server_vm.ip")

# Clone a repo
git clone ssh://git@$GIT_IP/srv/git/edge-cd.git
```

## Environment Variables

Customize test behavior:

```bash
# Override default artifact storage location
export E2E_ARTIFACTS_DIR=/custom/path

# VM Configuration (defaults: 2048MB memory, 2 VCPUs, 20GB disk)
export E2E_VM_MEMORY=4096
export E2E_VM_VCPUS=4
export E2E_DISK_SIZE=40G
```

## Cleanup

### List All Test Environments

```bash
edgectl-e2e list
```

Shows all persistent test environments with status.

### Remove a Specific Environment

```bash
edgectl-e2e teardown e2e-20231025-abc123
```

Destroys VMs, removes artifacts, updates metadata.

### Force Cleanup of Orphaned VMs

If tests were interrupted and VMs are still running:

```bash
# List running VMs
virsh list

# Manually destroy if needed
virsh destroy e2e-target-abc123
virsh undefine --remove-all-storage e2e-target-abc123
```

## Prerequisites

### qemu-kvm & libvirt

```bash
sudo apt update
sudo apt install qemu-kvm libvirt-daemon-system libvirt-dev pkg-config
sudo systemctl enable --now libvirtd
sudo usermod -a -G libvirt "${USER}"
# Verify group by running:
getent group | grep libvirt
# Test by running:
virsh --connect qemu:///system list --all
```

### qemu-img

```bash
sudo apt-get update
sudo apt-get install xorriso
```

### Additional Dependencies

```bash
sudo apt install libcap-ng-dev libseccomp-dev
cargo install virtiofsd
```

## Troubleshooting

### Test Hangs Waiting for VM IP

VMs sometimes take longer to get DHCP leases. Check:

```bash
# Connect to libvirt and check IP assignment
virsh net-dhcp-leases default

# If stuck, cancel test with Ctrl+C and cleanup
edgectl-e2e teardown <TEST_ID>
```

### SSH Connection Refused

Check that VM is actually running and SSH is available:

```bash
# Check VM status
virsh list

# Check VM console for errors
virsh console e2e-target-<ID>
```

### Git Server Not Reachable

Ensure git server VM is running and SSH is configured:

```bash
# Test git server SSH access
ssh git@<GIT_IP> "ls /srv/git"
```

## Virsh Debugging

Useful commands for debugging libvirt issues:

```bash
# Inspect default network configuration
virsh net-dumpxml default

# Release DHCP leases
virsh net-dhcp-release default
```

## Performance Notes

- Test takes 5-10 minutes total (mostly VM startup)
- First run downloads Ubuntu cloud image (takes longer)
- Subsequent runs reuse cached image
- With --keep-artifacts: VMs stay running, faster re-runs

## Clean Up All Test Resources

If you need to completely remove all test VMs and artifacts:

```bash
# Destroy all edge-cd VMs
for VM_NAME in $(virsh list --all | grep "e2e-" | awk '{print $2}'); do
    virsh destroy "${VM_NAME}"
    virsh undefine --remove-all-storage "${VM_NAME}"
done
# or 
{
    for VM_NAME in $(virsh list --all | awk 'NR > 2 {print $2}' | xargs); do
        virsh destroy "${VM_NAME}"
        sudo rm -rf "/tmp/${VM_NAME}*"
        virsh undefine --remove-all-storage "${VM_NAME}"
    done
}


# Remove all test artifacts
rm -rf ~/.edge-cd/e2e/
rm -rf /tmp/edge-cd-e2e-*
```
