#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

E2E_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
REPO_DIR="$(dirname "$(dirname "$(dirname "${E2E_DIR}")")")"

main() {
  docker build -t edge-cd-e2e -f "${E2E_DIR}/Dockerfile" "${REPO_DIR}"

docker run --rm -t edge-cd-e2e
}

main "$@"
