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
      echo "***********************************************************************************"
      echo "Checking: $test_dir"
      echo "***********************************************************************************"
      
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
  echo "Searching for modules to prepare..."

  # Find all directories containing a 'dagger.json'
  while IFS= read -r -d '' module_dir; do
  
    local module_name
    module_name="${module_dir##*/}"

    # Skip hidden folders like .dagger or test submodules
    if [[ "$module_name" == "tests" || "$module_name" == .* ]]; then
      continue
    fi

    echo " "
    echo "***********************************************************************************"
    echo "Prepare module: $module_name"
    echo "***********************************************************************************"
    
    # Prepare
    #   Run the command with plain logs, copying the output to the screen via tee
    #   while capturing the raw text string into our cmd_out variable
    set +e
    local cmd_out
    cmd_out=$("$0" prepare "$module_name" 2>&1 | tee /dev/stderr)
    exit_code="$?"
    set -e

    if [[ "$exit_code" -eq 0 ]]; then
      # Read the newly generated version from the file and save it
      local version
      version=$(cat "$module_name/VERSION")
      PREPARED_VERSIONS["$module_name"]="$version"
    else
      # The command returned an error code. Check the captured text variable for the bypass phrase.
      if [[ "$cmd_out" == *"there was nothing to bump"* ]]; then
          echo "[$module_name] Skipping '$module_name': No changes detected, there was nothing to bump."
      else
          # It's a completely different real failure! Crash out manually to protect the pipeline.
          echo "[$module_name] ERROR: '$0 prepare' failed with exit code $exit_code" >&2
          echo "$cmd_out" >&2  # Dump the error variable to stderr so you can see why it failed
          exit "$exit_code"
      fi
    fi

  # debug- hardcode list for faster testing, will need to revert!
  # done < <(find . -type f -name "dagger.json" -printf "%h\0")
  # done < <(printf "./sonarqube\0./shields/tests\0./shields\0./yamllint\0./markdownlint\0./.dagger")
  done < <(printf "./markdownlint\0./.dagger")


  # Summary of what was collected
  echo " "
  echo "***********************************************************************************"
  echo -e "\nSummary of Prepared Modules:"
  if [[ ${#PREPARED_VERSIONS[@]} -eq 0 ]]; then
    echo "No modules were bumped."
  else
    # Loop through the dictionary keys to show what was stored
    for mod in "${!PREPARED_VERSIONS[@]}"; do
      echo "  - $mod: ${PREPARED_VERSIONS[$mod]}"
    done

    echo -e "\nTODO:"
    echo -e "  - Review the local changes in each module listed above."
    echo -e "  - If all is good run: '$0 approve-all' to commit and tag each module\n"
    # Save the prepared versions for each module so that it can be read in and used by the approve-all
    mkdir -p .temp
    declare -p PREPARED_VERSIONS > .temp/PREPARED_VERSIONS.txt
  fi

}

# Run approve on all sub modules
approve_all() {
  echo "Searching for modules to approve..."

  if [ ! -f ".temp/PREPARED_VERSIONS.txt" ]; then
      echo "Did not find list of prepared module version.  File '.temp/PREPARED_VERSIONS.txt' does not exist. "
      echo "Did you forget to run '$0 prepare-all' first?"
      exit 1
  fi

  # read list of prepared version created by prepare-all
  source .temp/PREPARED_VERSIONS.txt
  rm .temp/PREPARED_VERSIONS.txt

  # Loop the list of prepared modules 
  for mod in "${!PREPARED_VERSIONS[@]}"; do

    ver="${PREPARED_VERSIONS[$mod]}"
    echo " "
    echo "***********************************************************************************"
    echo "Approve module: $mod  version: $ver"
    echo "***********************************************************************************"

    # Call approve for given module in a sub-shell
    ("$0" approve "$mod")

  done

  # Summary approved
  echo " "
  echo "***********************************************************************************"
  echo -e "\nSummary of Approved Modules:"
  if [[ ${#PREPARED_VERSIONS[@]} -eq 0 ]]; then
    echo "No modules were approved."
  else
    # Loop through the dictionary keys to show what was stored
    for mod in "${!PREPARED_VERSIONS[@]}"; do
      echo "  - $mod: ${PREPARED_VERSIONS[$mod]}"
    done

    echo -e "\nTODO:"
    echo -e "  - Review the local changes in each module listed above."
    echo -e "  - If all is good run: '$0 approve-all' to commit and tag each module\n"
    # Save the prepared versions for each module so that it can be read in and used by the approve-all
    mkdir -p .temp
    declare -p PREPARED_VERSIONS > .temp/PREPARED_VERSIONS.txt
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

    echo "Successfully ran 'git add/commit/tag', to release run: $0 publish $module"
    # if confirm_continue publish; then
    #   "$0" publish "$module"
    # fi

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
approve-all)
    echo "Running approve for all modules..."
    approve_all
    ;;

*)
    help
    ;;
esac
