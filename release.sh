#!/usr/bin/env bash
set -euo pipefail
source ./update-deps.sh
# Required env vars:
# GITHUB_TOKEN - github repo api access


cmd=$1
shift

module=""

# Map were key=module name, value=version 
declare -A PREPARED_VERSIONS=()

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
check_all() {
  local any_failed=0

  while IFS= read -r -d '' test_dir; do
    if [[ -f "$test_dir/dagger.json" ]]; then
      echo "Checking: $test_dir"
      
      # If the subshell fails, set the failure flag to 1
      if ! (cd "$test_dir" && dagger check); then
        echo "Check failed in $test_dir"
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

# Run prepare on all sub modules
prepare_all() {
  echo "Searching for modules to prepare for release..."

  # Find all directories containing a 'dagger.json'
  while IFS= read -r -d '' module_dir; do
  
    local module_name
    module_name=$(basename "$(dirname "$module_dir")")

    # Skip hidden folders like .dagger or test submodules
    if [[ "$module_name" == "tests" || "$module_name" == .* ]]; then
      continue
    fi

    echo "Process module: $module_name"
    git fetch --tags
    
    # Run module tests only if the tests subdirectory exists
    if [[ -d "$module_name/tests" ]]; then
      echo "[$module_name] Run checks for module: $module_name"
      if ! dagger -m "$module_name/tests" checks; then
        echo "[$module_name] Error: Tests failed for module '$module_name'!"
        exit 1
      fi
    else
      echo "[$module_name] No tests directory found for '$module_name', skipping tests."
    fi

    # Capture stdout and stderr into a variable to inspect it
    echo "[$module_name] Run prepare for module: $module_name"
    local output
    if ! output=$(dagger call --module="$module_name" prepare 2>&1); then
      
      # Check if the failure was simply because there was nothing to bump
      if echo "$output" | grep -q "there was nothing to bump"; then
        echo "[$module_name] Skipping '$module_name': No changes detected to bump."
      else
        echo "[$module_name] Error: 'dagger prepare' failed for module '$module_name'!"
        echo "----------------- DAGGER OUTPUT -----------------"
        echo "$output"
        echo "-------------------------------------------------"
        exit 1
      fi
    else
      echo "$output"
      echo "[$module_name] Successfully prepared release for '$module_name'."

      # -----------------------------------------------------------------
      # NEW DICTIONARY LOGIC:
      # Read the newly generated version from the file and save it
      # -----------------------------------------------------------------
      local version
      version=$(cat "$module_name/VERSION")
      PREPARED_VERSIONS["$module_name"]="$version"
      
      echo "Tracked version: $module_name -> $version"
    fi

  done < <(find . -type f -name "dagger.json" -exec dirname {} \; -print0)

  # Summary of what was collected
  echo "=================================================="
  echo "Summary of Prepared Modules:"
  echo "=================================================="
  if [[ ${#PREPARED_VERSIONS[@]} -eq 0 ]]; then
    echo "No modules were bumped."
  else
    # Loop through the dictionary keys to show what was stored
    for mod in "${!PREPARED_VERSIONS[@]}"; do
      echo "  • $mod: v${PREPARED_VERSIONS[$mod]}"
    done
    echo "=================================================="
    
    # Optional prompt to approve all collected changes
    if confirm_continue "approve all changes"; then
        for mod in "${!PREPARED_VERSIONS[@]}"; do
          "$0" approve "$mod"
        done
    fi
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

    # Run module tests only if the tests subdirectory exists
    if [[ -d "$module_name/tests" ]]; then
      dagger -m "$module_name/tests" checks
    else
      echo "ℹNo tests directory found for '$module_name', skipping tests."
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
    check_all
    ;;

prepare-all)
    echo "Running prepare for all modules..."
    prepare_all
    ;;

*)
    help
    ;;
esac
