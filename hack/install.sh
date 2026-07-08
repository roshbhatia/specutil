#!/usr/bin/env bash
set -euo pipefail

main() {
  local repo_root
  repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

  local bin_dir
  bin_dir="${BIN_DIR:-${HOME}/.local/bin}"

  if ! command -v go > /dev/null 2>&1; then
    printf 'error: go is required on PATH\n' >&2
    return 1
  fi

  mkdir -p "${bin_dir}"
  (cd "${repo_root}" && GOBIN="${bin_dir}" go install ./cmd/specutil)

  printf 'installed specutil to %s/specutil\n' "${bin_dir}"
  case ":${PATH}:" in
    *":${bin_dir}:"*) ;;
    *) printf 'warning: %s is not on PATH\n' "${bin_dir}" >&2 ;;
  esac
}

main "$@"
