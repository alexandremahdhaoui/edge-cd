# edgectl Architecture Decisions

This document outlines key architectural decisions for the `edgectl` CLI tool, specifically focusing on the `bootstrap` command.

## 1. Idempotency Strategy

**Decision**: `edgectl bootstrap` will primarily rely on **pre-flight checks** for each operation to ensure idempotency.

**Explanation**:
Instead of maintaining a remote state file (like the `bootstrap-router` shell scripts), each step in the `bootstrap` process will first check the current state of the remote device. For example:
- Before installing a package, it will check if the package is already installed.
- Before cloning a repository, it will check if the repository directory already exists. If it exists, it will attempt to pull updates instead of cloning again.
- Before enabling a service, it will check if the service is already enabled.

This approach simplifies the state management on the remote device and makes each operation self-contained and repeatable without side effects if run multiple times.

## 2. Unit Testing Strategy

**Decision**: Unit tests will mandate the use of **Go interfaces for all external-facing components** (e.g., SSH client, filesystem operations, Docker commands). Unit tests will then use **simple, hand-written mocks** of these interfaces.

**Explanation**:
-   **Interfaces**: Defining interfaces for external dependencies allows for easy substitution of real implementations with mock implementations during unit testing. This decouples the business logic from the underlying infrastructure.
-   **Hand-written Mocks**: For simplicity and control, mocks will be implemented manually within the test files or dedicated mock files. This avoids the overhead and complexity of external mocking frameworks for this project's scope.
-   **E2E Tests**: Real implementations of these interfaces will be used in End-to-End (E2E) tests to verify the integration with actual external systems (e.g., a real SSH connection to a Docker container).

## 3. Error Handling and Transactionality Strategy

**Decision**: The application will adopt a **fail-fast** approach with explicit error propagation. For critical operations, a **best-effort cleanup** will be attempted, but full transactionality (rollback) will not be implemented in the initial version for simplicity.

**Explanation**:
-   **Fail-Fast**: Upon encountering a non-recoverable error during any step of the bootstrap process, the `edgectl` command will immediately terminate and report the error. This prevents the device from being left in an inconsistent or partially configured state without explicit notification.
-   **Error Propagation**: Errors will be returned from functions and handled at appropriate levels, typically leading to a `t.Fatalf` in tests or an `os.Exit(1)` in the main application.
-   **Best-Effort Cleanup**: In case of failure, the `t.Cleanup` mechanism in tests (and potentially `defer` statements in the main application) will be used to attempt to revert or clean up any changes made *up to the point of failure*. However, a full, guaranteed transactional rollback is complex and will be deferred for future consideration if required by more stringent requirements. The idempotency strategy (pre-flight checks) also aids in recovery from partial failures, as a re-run of the command should pick up where it left off.