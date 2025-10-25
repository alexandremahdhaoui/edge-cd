# edgectl-e2e - E2E Test Environment Manager

CLI tool for managing e2e test environments independently of the test harness.

## Installation

```bash
go build -o ~/bin/edgectl-e2e ./cmd/edgectl-e2e
```

## Usage

### Commands

#### create

Create and provision a new e2e test environment.

```bash
edgectl-e2e create
```

**Output:**
- Environment ID (to stdout, for scripting)
- Target VM name and IP
- Target VM SSH command: `ssh -i <key> ubuntu@<ip>`
- Git server VM name and IP
- Git server SSH command: `ssh -i <key> git@<ip>`
- Artifact directory path
- Git repository URLs and SSH access commands
- Temp directory structure

#### get

Get detailed information about a test environment.

```bash
edgectl-e2e get <test-id>
```

**Output:**
- Environment status and creation time
- Target VM: name, IP, SSH command, memory, vCPUs
- Git server VM: name, IP, SSH command, memory, vCPUs
- Git repositories with SSH URLs
- SSH key file paths
- Temp directory structure

#### run

Execute e2e tests in an existing environment.

```bash
edgectl-e2e run <test-id>
```

**Example:**
```bash
edgectl-e2e run e2e-20231025-abc123
```

**Output:**
- Test progress
- Test results summary
- Pass/fail status

#### delete

Destroy a test environment and clean up all resources.

```bash
edgectl-e2e delete <test-id>
```

**What gets cleaned up:**
- Target VM (destroyed in libvirt)
- Git server VM (destroyed in libvirt)
- Entire temp directory tree: `/tmp/e2e-<test-id>/`
- Artifact directory
- Metadata in artifact store

#### test

One-shot test: create → run → delete.

```bash
edgectl-e2e test
```

Equivalent to running `go test ./test/edgectl/e2e` directly.

#### list

Show all test environments and their status.

```bash
edgectl-e2e list
```

**Output:** Table showing:
- Environment ID
- Creation time
- Current status (created/running/passed/failed)
- Target and git server VM names

### Exit Codes

- 0 = Success
- 1 = Error (wrong arguments, operation failed, etc.)

### Temporary Directory Structure

Each test environment creates a structured temporary directory at `/tmp/e2e-<test-id>/` containing:

```
/tmp/e2e-<test-id>/
├── vmm/              # VM disk overlays and cloud-init ISOs
├── gitserver/        # Git server artifacts and VM files
└── artifacts/        # Test artifacts and logs (local, not used by current setup)
```

All subdirectories are owned by the test environment and deleted on cleanup with a single `os.RemoveAll()` call.

## Examples

### Basic Workflow

```bash
# Create environment
edgectl-e2e create
# e2e-20231025-abc123

# Get environment details with SSH commands
edgectl-e2e get e2e-20231025-abc123

# Run tests
edgectl-e2e run e2e-20231025-abc123

# Cleanup (deletes /tmp/e2e-20231025-abc123/)
edgectl-e2e delete e2e-20231025-abc123
```

### One-Shot Test

```bash
# Equivalent to: go test ./test/edgectl/e2e
edgectl-e2e test
```

### List and Cleanup Old Tests

```bash
# See all environments
edgectl-e2e list

# Get info about a specific environment
edgectl-e2e get e2e-20231025-abc123

# Cleanup old environment
edgectl-e2e delete e2e-20231024-old123
```

### Manual VM Access Between Tests

```bash
# Create environment and capture the ID
ENV_ID=$(edgectl-e2e create)

# Get SSH commands to VMs
edgectl-e2e get $ENV_ID

# SSH into target VM for inspection
ssh -i /path/to/key ubuntu@<TARGET_IP>

# SSH into git server for inspection
ssh -i /path/to/key git@<GIT_SERVER_IP>

# Run tests when ready
edgectl-e2e run $ENV_ID

# Cleanup
edgectl-e2e delete $ENV_ID
```

## Environment Variables

### E2E_ARTIFACTS_DIR

Override the default artifact storage location.

```bash
export E2E_ARTIFACTS_DIR=/custom/path
edgectl-e2e setup
```

Default: `~/.edge-cd/e2e/`

## Advanced Usage

### Inspect VMs Between Tests

```bash
# Create environment and save ID
ENV_ID=$(edgectl-e2e create)

# Get SSH commands
edgectl-e2e get $ENV_ID

# SSH into target VM for inspection
ssh -i <KEY_PATH> ubuntu@<TARGET_IP>

# Run tests when ready
edgectl-e2e run $ENV_ID

# Cleanup
edgectl-e2e delete $ENV_ID
```

### Recovery from Failed Test

```bash
# If a test fails and you want to keep the environment:
ENV_ID=$(edgectl-e2e create)

edgectl-e2e run $ENV_ID
# Test fails, but VMs still running

# Get SSH access info
edgectl-e2e get $ENV_ID

# Inspect and debug
ssh -i <KEY_PATH> ubuntu@<TARGET_IP>

# When done debugging, cleanup
edgectl-e2e delete $ENV_ID
```

### List Environments and Filter

```bash
# List all
edgectl-e2e list

# List only passed environments (using grep)
edgectl-e2e list | grep passed

# Get details for a specific environment
edgectl-e2e get e2e-20231025-abc123
```

## Troubleshooting

### "environment not found" error

```bash
# Verify the test ID is correct
edgectl-e2e list

# Use the exact ID from the list
edgectl-e2e run e2e-20231025-abc123
```

### Create Hangs or Fails

Check libvirt status:

```bash
# Verify libvirt is running
sudo systemctl status libvirtd

# Check if you're in the libvirt group
groups | grep libvirt

# If not, add yourself and login again
sudo usermod -a -G libvirt "${USER}"
```

### Can't SSH into VM

```bash
# Get SSH command from environment
edgectl-e2e get e2e-20231025-abc123

# Find your SSH key path from output and verify it exists
ls -la <SSH_KEY_PATH>

# Check permissions (should be 0600)
chmod 0600 <SSH_KEY_PATH>

# Try connecting with verbose output
ssh -v -i <SSH_KEY_PATH> ubuntu@<VM_IP>
```

### Delete Fails with "VM not found"

If the VM was already destroyed or cleaned up manually:

```bash
# Delete operation is best-effort - it will remove metadata and temp dirs
edgectl-e2e delete <test-id>

# If metadata is stuck, check manually
rm ~/.edge-cd/e2e/artifacts.json

# Clean up orphaned temp directories if needed
rm -rf /tmp/e2e-*
```

### Temp Directory Not Cleaned Up

All temporary files are located in `/tmp/e2e-<test-id>/`. To manually clean up:

```bash
# List all temp directories
ls -la /tmp/e2e-*

# Remove a specific test's temp directory
rm -rf /tmp/e2e-20231025-abc123/
```

## See Also

- `test/edgectl/e2e/README.md` - E2E testing guide
- `go test ./test/edgectl/e2e -v` - Run tests programmatically
- `virsh list --all` - List all VMs
- `virsh console <VM_NAME>` - Access VM console
