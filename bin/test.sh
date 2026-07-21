#!/usr/bin/env bash

ensure_in_project_root() {
  local script_dir
  local project_root

  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  project_root="$(cd "${script_dir}/.." && pwd)"

  if [ "$(pwd)" != "${project_root}" ]; then
    echo "Error: This script must be run from the project root (${project_root})." >&2
    exit 1
  fi
}

# Ensure user is in project root
ensure_in_project_root

printf "all good\n"
