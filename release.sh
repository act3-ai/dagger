#!/usr/bin/env bash
set -euo pipefail
source ./update-deps.sh
# Required env vars:
# GITHUB_TOKEN - github repo api access


cmd=$1
shift

module=""

confirm_continue() {
  local next_step="$1"

  read -r -p "Continue to '$next_step'? [y/N] " reply
  case "$reply" in
    [yY]) return 0 ;;
    *) return 1 ;;
  esac
}

# Check if module arg is set 
require_module() {
  if [[ -z "$module" ]]; then
    echo "Error: The '$cmd' command requires a module argument."
    exit 1
  fi
}

# Run all checks in sub modules
run_dagger_checks_all() {
  local any_failed=0

  while IFS= read -r -d '' test_dir; do
    if [[ -f "$test_dir/dagger.json" ]]; then
      echo "Checking: $test_dir"
      
      # If the subshell fails, set the failure flag to 1
      if ! (cd "$test_dir" && dagger check); then
        echo "❌ Check failed in $test_dir"
        any_failed=1
      fi
    fi
  done < <(find . -type d -name "tests" -print0)

  # If any check failed, exit the parent process with an error code now
  if [[ $any_failed -eq 1 ]]; then
    echo "One or more Dagger checks failed."
    exit 1
  fi
}

# Loop through remaining args
while [[ $# -gt 0 ]]; do
  case "$1" in
    -*)
      echo "Unknown option: $1"
      exit 1
      ;;
    *)
      # First non-flag argument is module
      if [[ -z "$module" ]]; then
        module=$1
        shift
      else
        echo "Unexpected argument: $1"
        exit 1
      fi
      ;;
  esac
done

case "$cmd" in
prepare)
    # Enforces that $module is present
    require_module

    git fetch --tags

    #run module tests
    if [[ "$module" == "govulncheck" || "$module" == "renovate" || "$module" == "sonarqube" ]]; then
      : # Do nothing
    else
      dagger -m "$module/tests" checks
    fi

    dagger call --module="$module" prepare
    version=$(cat "$module/VERSION")

    echo "Please review the local changes, especially $module/releases/$version.md"
    if confirm_continue approve; then
      "$0" approve "$module"
    fi

    ;;

approve)
    # Enforces that $module is present
    require_module

    version=$(cat "$module/VERSION")

    notesPath="$module/releases/v$version.md"
    # release material
    git add "$module/VERSION" "$module/CHANGELOG.md" "$notesPath"
    # signed commit
    git commit -S -m "chore(release): prepare for $module/v$version"
    # annotated and signed tag
    git tag -s -a -m "Official release $module/v$version" "$module/v$version"

    if confirm_continue publish; then
      "$0" publish "$module"
    fi

    ;;
publish)
    # Enforces that $module is present
    require_module

    # push this branch and the associated tags
    git push --follow-tags

    version=$(cat "$module/VERSION")

    dagger call --module="$module" release --version="$version"

    ;;

check-all)
    echo "Running all checks for all modules..."
    run_dagger_checks_all
    ;;

*)
    help
    ;;
esac
