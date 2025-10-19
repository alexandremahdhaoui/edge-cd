# edge-cd.sh refactoring

## Goals

| Goals                         | Status |
|-------------------------------|--------|
| Clone repo to /usr/local/     | `OPEN` |
| Support lib files             | `OPEN` |
| Add proper logging            | `OPEN` |
| Add locking                   | `OPEN` |
| Persist config in /etc/       | `OPEN` |
| Allow configuring variables   | `OPEN` |
| Step-by-step update mode      | `OPEN` |
|                               | `OPEN` |

##### Clone repo to /usr/local

Clone this repo or do a sparse checkout

##### Support lib files

Require to clone the `edge-cd` repo.

##### Add proper logging

Add a `log_info` and `log_err` func.

##### Add locking

Ensure `edge-cd.sh` cannot run in parallel by implemnting a simple lock in /tmp/edge-cd.lock
Write PID into file, so that if process exit without removing lock, a new process can take over.

##### Persist user config in /etc? Or

Move the config or create a symlink thats move

##### Step-by-step update mode

This mode would ensure that each changes are applied commit by commit
in order.
