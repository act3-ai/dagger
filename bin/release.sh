#!/usr/bin/env bash
set -euo pipefail
# Required env vars:
# GITHUB_TOKEN - github repo api access

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

confirm_continue() {
  local next_step="$1"

  echo " "
  read -r -p "Continue to '$next_step'? [y/N] " reply
  case "$reply" in
    [yY]) return 0 ;;
    *) return 1 ;;
  esac
}

require_one_module() {
  # Check if exactly one module arg is set 
  if [[ ${#modules[@]} -ne 1 ]]; then
    echo "Error: The '$cmd' command requires exactly one module argument. (You provided ${#modules[@]})"
    exit 1
  fi  
}

require_some_module() {
  # Check if at least one module (aka "some") arg is set 
  if [[ ${#modules[@]} -lt 1 ]]; then
    echo "Error: The '$cmd' command requires one or more module arguments. (You provided ${#modules[@]})"
    exit 1
  fi  
}

require_no_module() {
  # Check to ensure NO module args are set 
  if [[ ${#modules[@]} -ne 0 ]]; then
    echo "Error: The '$cmd' should not have any module arguments. (You provided ${#modules[@]})"
    exit 1
  fi    
}

# Checks if a module has untagged git changes.
# Returns 0 if changed/new, 1 if up-to-date.
has_module_changed() {
  local mod="$1"
  local current_tag="$2"

  # If tag doesn't exist, treat it as changed (first release)
  if ! git rev-parse --verify --quiet "$current_tag" >/dev/null 2>&1; then
    echo "UNKNOWN: Tag '$current_tag' does not exist for '$mod' (first release?)"
    return 0
  fi

  # Check if files inside the module folder changed since the tag
  if ! git diff --quiet "$current_tag" HEAD -- "$mod"; then
    echo "CHANGED: '$mod' has changes since tag '$current_tag'"
    return 0
  else
    echo "Up-to-date: '$mod' has no changes since tag '$current_tag'"
    return 1
  fi
}

# Run all checks in sub modules
check_all() {
  local any_failed=0
  local test_dir

  for dir in */; do
    test_dir="${dir}tests"
    if [[ ! -f "$test_dir/dagger.json" ]]; then
      continue
    fi

    echo "Checking: $test_dir"
    
    # If the subshell fails, set the failure flag to 1
    if ! (cd "$test_dir" && dagger check); then
      echo "Check failed in $test_dir"
      any_failed=1
    fi

  done

  # If any check failed, exit the parent process with an error code now
  if [[ $any_failed -eq 1 ]]; then
    echo "One or more Dagger checks failed."
    exit 1
  fi
}

# Run prepare on all sub modules
prepare_all() {
  require_no_module
  git fetch --tags

  # Declare an empty indexed array to hold prepared modules
  local -a prepared_modules=()

  # Loop through all immediate subdirectories
  for dir in */; do
  
    # Remove trailing slash
    mod="${dir%/}"
    version_file="$mod/VERSION"
    
    # folders that we want to skip
    if [[ "$mod" =~ ^(bin|\.dagger)$ ]]; then
      continue
    fi    

    # version file is required
    if [[ ! -f "$version_file" ]]; then
      printf "ERROR, version file: %s for module: %s not found\n" "$version_file" "$mod"
      exit 1
    fi

    # Read version and construct the expected tag name
    local version
    version=$(tr -d '[:space:]' < "$version_file")
    local current_tag="${mod}/v${version}"

    # filter out unchanged modules
    has_module_changed "$mod" "$current_tag" || continue

    ####################################
    # Prepare (Only reached if changed)
    ####################################
    echo "Prepare module: $mod"
    exit_code="0" # assume all good

    if [[ "$dry_run" == "true" ]]; then
      printf "[DRY-RUN] Would have called: %s prepare %s\n" "$0" "$mod"
    else
      #   Run the command with plain logs, copying the output to the screen via tee
      #   while capturing the raw text string into our cmd_out variable
      set +e
      local cmd_out
      cmd_out=$(BATCH_MODE=true "$0" "prepare" "$mod" 2>&1 | tee /dev/stderr)
      exit_code="$?"
      set -e
    fi

    # Handle completion logic...
    if [[ "$exit_code" -eq 0 ]]; then
      # Only consider the module prepared if there is something to bump
      prepared_modules+=("$mod")
    else
      # The command returned an error code. Check the captured text variable for the bypass phrase.
      if [[ "$cmd_out" == *"there was nothing to bump"* ]]; then
          echo "[$mod] Skipping '$mod': No changes detected, there was nothing to bump."
      else
          # It's a completely different real failure! Crash out manually to protect the pipeline.
          echo "[$mod] ERROR: '$0 prepare' failed with exit code $exit_code" >&2
          echo "$cmd_out" >&2  # Dump the error variable to stderr so you can see why it failed
          exit "$exit_code"
      fi
    fi

  done

  ####################################
  # Summary of what was collected
  ####################################
  printf "\nSummary of Prepared Modules:\n"
  if [[ ${#prepared_modules[@]} -eq 0 ]]; then
    echo "No modules were bumped."
    return 0
  fi

  # Print list of prepared modules
  for mod in "${prepared_modules[@]}"; do
    echo "  - $mod: $(cat "$mod/VERSION")"
  done

  # Ask user to continue
  if confirm_continue approve-all; then
    "$0" "approve-all" "${prepared_modules[@]}" ${dry_run:+"--dry-run"}
  fi 

}

# Run approve on all prepared modules
approve_all() {
  require_some_module

  # for each module that was prepared
  for mod in "${modules[@]}"; do
    echo "Approve module: $mod"
    if [[ "$dry_run" == "true" ]]; then
      printf "[DRY-RUN] Would have called: %s approve %s\n" "$0" "$mod"
    else
      # Call approve for given module in a sub-shell
      (BATCH_MODE=true "$0" approve "$mod")
    fi
  done

  # Summary of approved
  printf "\nSummary of Approved Modules:\n"
  for mod in "${modules[@]}"; do
    echo "  - $mod: $(cat "$mod/VERSION")"
  done

  # Ask user to continue
  if confirm_continue publish-all; then
    "$0" "publish-all" "${modules[@]}" ${dry_run:+"--dry-run"}
  fi 

}


# Run publish on all approved modules
publish_all() {
  require_some_module

  # for each module that was approved
  for mod in "${modules[@]}"; do
    ver="$(cat "$mod/VERSION")"
    echo "Publish module: $mod  version: $ver"

    if [[ "$dry_run" == "true" ]]; then
      printf "[DRY-RUN] Would have called: %s publish %s\n" "$0" "$mod"
    else
      # Call publish for given module in a sub-shell
      (BATCH_MODE=true "$0" publish "$mod")
    fi
  done

  # Summary of approved
  printf "\nSummary of Published Modules:\n"
  for mod in "${modules[@]}"; do
    echo "  - $mod: $(cat "$mod/VERSION")"
  done

  printf "\nDone."

}

prepare() {
    # Enforces that $module is present
    require_one_module

    # Use first module in the array
    local module="${modules[0]}"

    git fetch --tags

    # Run module tests only if the tests subdirectory exists
    if [[ -d "$module/tests" ]]; then
      dagger -m "$module/tests" checks
    else
      echo "No tests directory found for '$module', skipping tests."
    fi

    dagger call --auto-apply --progress=dots --module="$module" prepare

    # Skip prompt if in batch mode
    version=$(cat "$module/VERSION")
    if [[ "$batch_mode" == "true" ]]; then
        echo "Skip prompt, running in batch mode"
        return 0
    fi

    # Prompt user to continue to next step
    echo "Please review the local changes, especially $module/releases/$version.md"
    if confirm_continue approve; then
      "$0" approve "$module"
    fi    
    
}

approve() {
    # Enforces that $module is present
    require_one_module

    # Use first module in the array
    local module="${modules[0]}"

    version=$(cat "$module/VERSION")

    notesPath="$module/releases/v$version.md"
    # release material
    git add "$module/VERSION" "$module/CHANGELOG.md" "$notesPath"
    # signed commit
    git commit -S -m "chore(release): prepare for $module/v$version"
    # annotated and signed tag
    git tag -s -a -m "Official release $module/v$version" "$module/v$version"

    # Skip prompt if in batch mode
    if [[ "$batch_mode" == "true" ]]; then
        echo "Skip prompt, running in batch mode"
        return 0
    fi

    # Prompt user to continue to next step
    echo "Please review the local changes, especially $module/releases/$version.md"
    if confirm_continue publish; then
      "$0" approve "$module"
    fi    

}

publish() {
    # Enforces that $module is present
    require_one_module

    # Use first module in the array
    local module="${modules[0]}"

    # push this branch and the associated tags
    git push --follow-tags

    version=$(cat "$module/VERSION")

    dagger call --module="$module" release --version="$version"

    echo "Successfully ran 'dagger release' on module $module"
}

################################################################################################
#   Main
################################################################################################

# Ensure user is in project root
ensure_in_project_root

# Initialize an empty array to store multiple modules
modules=()

# Initialize the dry-run flag tracking variable
dry_run="false"

# The _all functions will set this to true to control the lower lever prompting
# an external env var BATCH_MODE is use to span process boundaries
batch_mode="${BATCH_MODE:-}"

cmd=$1
shift

# Loop through remaining args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      dry_run="true"
      shift
      ;;  
    -*)
      echo "Unknown option: $1"
      exit 1
      ;;
    *)
      # Every non-flag argument is added to the modules array
      modules+=("$1")
      shift
      ;;
  esac
done

case "$cmd" in
prepare)
    printf "\nRunning prepare..."
    prepare
    ;;

approve)
    printf "\nRunning approve..."
    approve
    ;;

publish)
    printf "Running publish..."
    publish
    ;;

check-all)
    printf "\nRunning all checks for all modules...\n"
    check_all    
    ;;

prepare-all)
    printf "\nRunning prepare for all modules...\n"
    prepare_all   
    ;;

approve-all)
    printf "\nRunning approve for all prepared modules...\n"
    approve_all 
    ;;

publish-all)
    printf "\nRunning publish for all approved modules...\n"
    publish_all
    ;;

*)
    help
    ;;
esac
