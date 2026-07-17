#!/usr/bin/env bash
set -euo pipefail
source ./update-deps.sh
# Required env vars:
# GITHUB_TOKEN - github repo api access

confirm_continue() {
  local next_step="$1"

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
  require_no_module
  
  git fetch --tags

  # Declare an empty indexed array to hold prepared modules
  local -a prepared_modules=()

  echo "Searching for modules to prepare..."
  # Loop through all immediate subdirectories
  for dir in */; do
  
    # Remove trailing slash
    mod="${dir%/}"
    version_file="$mod/VERSION"

    # Skip if the directory doesn't contain a VERSION file
    if [ ! -f "$version_file" ]; then
        continue
    fi

    # Read version and construct the expected tag name
    version=$(tr -d '[:space:]' < "$version_file")
    current_tag="${mod}/v${version}"

    echo -e "\nPrepare module: $mod"

    # Verify if the computed tag actually exists in the local git database
    if git rev-parse --verify --quiet "$current_tag" >/dev/null 2>&1; then
        # Check if any tracked files in this directory changed since the tag
        if ! git diff --quiet "$current_tag" HEAD -- "$mod"; then
            echo "CHANGED: '$mod' has changes since tag '$current_tag'"
        else
            echo "Up-to-date: '$mod' has no changes since tag '$current_tag'"
            continue
        fi
    else
        echo "UNKNOWN: Tag '$current_tag' does not exist for '$mod' (first release?)"
        continue
    fi

    # Prepare
    #   Run the command with plain logs, copying the output to the screen via tee
    #   while capturing the raw text string into our cmd_out variable
    
    exit_code="0" # assume all good

    if [[ "$dry_run" == "true" ]]; then
      echo "[DRY-RUN] Would have called:"
      echo "$0 prepare $mod"
    else
      set +e
      local cmd_out
      cmd_out=$("$0" "prepare" "$mod" 2>&1 | tee /dev/stderr)
      exit_code="$?"
      set -e
    fi

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


  # Summary of what was collected
  echo -e "\nSummary of Prepared Modules:"
  if [[ ${#prepared_modules[@]} -eq 0 ]]; then
    echo "No modules were bumped."
    return 0
  fi

  # Print list of prepared modules
  for mod in "${prepared_modules[@]}"; do
    echo "- $mod"
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
    echo -e "\nApprove module: $mod"
    if [[ "$dry_run" == "true" ]]; then
      echo "[DRY-RUN] Would have called:"
      echo "$0 approve $mod"
    else
      # Call approve for given module in a sub-shell
      ("$0" approve "$mod")
    fi
  done

  # Summary of approved
  echo -e "\nSummary of Approved Modules:"
  for mod in "${modules[@]}"; do
    echo "  - $mod: $(cat "$mod/VERSION")"
  done

  # Ask user to continue
  if confirm_continue publish-all; then
    "$0" "publish-all" "${mod[@]}" ${dry_run:+"--dry-run"}
  fi 

}


# Run publish on all approved modules
publish_all() {
  require_some_module

  # for each module that was approved
  for mod in "${modules[@]}"; do
    ver="$(cat "$mod/VERSION")"
    echo -e "\nPublish module: $mod  version: $ver"

    if [[ "$dry_run" == "true" ]]; then
      echo "[DRY-RUN] Would have called:"
      echo "$0 publish $mod"
    else
      # Call publish for given module in a sub-shell
      ("$0" publish "$mod")
    fi
  done

  # Summary of approved
  echo -e "\nSummary of Published Modules:"
  for mod in "${modules[@]}"; do
    echo "  - $mod: $(cat "$mod/VERSION")"
  done

  echo -e "\nDone."

}

prepare() {
    # Enforces that $module is present
    require_one_module

    # Use first module in the array
    local mod
    module="${modules[0]}"

    git fetch --tags

    # Run module tests only if the tests subdirectory exists
    if [[ -d "$module/tests" ]]; then
      dagger -m "$module/tests" checks
    else
      echo "No tests directory found for '$module', skipping tests."
    fi

    dagger call --auto-apply --progress=dots --module="$module" prepare
    version=$(cat "$module/VERSION")
    echo -e "\n NOTE:"
    echo "  - Please review the local changes, especially $module/releases/$version.md"
    echo "  - If all is good run: $0 approve $module"
}

approve() {
    # Enforces that $module is present
    require_one_module

    # Use first module in the array
    local mod
    module="${modules[0]}"

    version=$(cat "$module/VERSION")

    notesPath="$module/releases/v$version.md"
    # release material
    git add "$module/VERSION" "$module/CHANGELOG.md" "$notesPath"
    # signed commit
    git commit -S -m "chore(release): prepare for $module/v$version"
    # annotated and signed tag
    git tag -s -a -m "Official release $module/v$version" "$module/v$version"

    echo "Successfully ran 'git add/commit/tag', to release run: $0 publish $module"

}

publish() {
    # Enforces that $module is present
    require_one_module

    # Use first module in the array
    local mod
    module="${modules[0]}"

    # push this branch and the associated tags
    git push --follow-tags

    version=$(cat "$module/VERSION")

    dagger call --module="$module" release --version="$version"

    echo "Successfully ran 'dagger release' on module $module"
}

################################################################################################
#   Main
################################################################################################

# Initialize an empty array to store multiple modules
modules=()

# Initialize the dry-run flag tracking variable
dry_run="false"

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
    echo "Running prepare..."
    prepare
    ;;

approve)
    echo "Running approve..."
    approve
    ;;

publish)
    echo "Running publish..."
    publish
    ;;

check-all)
    echo "Running all checks for all modules..."
    check_all    
    ;;

prepare-all)
    echo "Running prepare for all modules..."
    prepare_all   
    ;;

approve-all)
    echo "Running approve for all prepared modules..."
    approve_all 
    ;;

publish-all)
    echo "Running publish for all approved modules..."
    publish_all
    ;;

*)
    help
    ;;
esac
