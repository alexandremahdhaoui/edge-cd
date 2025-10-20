# TODOs

- FIX UNIT TESTS
- Add E2E test target in Makefile

## Dashboard

#### Open tasks

| Tasks                              | Status   |
|------------------------------------|----------|
| [003] Update examples/config.yaml  | `OPEN`   |
| [004] Implement unit tests         | `OPEN`   |
| [005] Implement test w/ container  | `OPEN`   |
| [002] Create edgectl bootstrap CLI | `OPEN`   |
| [001] Add support for systemd      | `OPEN`   |
| [006] Step-by-step update mode     | `OPEN`   |

#### Closed tasks

| Tasks                              | Status   |
|------------------------------------|----------|
| [000] Implement service manager    | `CLOSED` |
| Clone repo to /usr/local/          | `CLOSED` |
| Support lib files                  | `CLOSED` |
| Add proper logging                 | `CLOSED` |
| Add locking                        | `CLOSED` |
| Persist config in /etc/            | `CLOSED` |
| Allow configuring variables        | `CLOSED` |

## [002] Step-by-step update mode

This mode ensures that each commits are applied atomically in their sequence.

Requires:

- Monitoring
- Rollback mechanism
  - which basically involves rolling out the failing commit and marking it as unsafe
  - it requires being able to pin a specific commit (or just create a list of banned commits)
  - how do we deal with package being installed/upgraded?
