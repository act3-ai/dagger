#!/usr/bin/env bash

list_modules() {
  find . -maxdepth 2 -mindepth 2 -type f -name dagger.json -exec dirname {} \; | sed 's|^\./||'
}

list_modules_with_tests() {
  find . -maxdepth 3 -type f -path './*/tests/dagger.json' -exec dirname {} \; | sed 's|^\./||'
}

detect_latest_dagger_version() {
  curl -s https://api.github.com/repos/dagger/dagger/releases/latest | jq -r '.tag_name'
}



check_git_status() {
  local path="${1:-}"

  if [[ -z "$path" ]]; then
    # No argument provided: check entire repo
    git status --porcelain
  else
    # Argument provided: check only the given path
    git status --porcelain "$path"
  fi
}

upgrade_dagger_engine() {
  if [[ -z "$1" ]]; then
    #Check if module name given
    echo "Error: No module name given to upgrade"
    exit 1
  fi

  module="$1"
  LATEST_DAGGER_VERSION=$(detect_latest_dagger_version)
  CURRENT_DAGGER_VERSION=$(jq -r '.engineVersion' "$1/dagger.json")

  if [[ "$CURRENT_DAGGER_VERSION" != "$LATEST_DAGGER_VERSION" ]]; then
    echo "Upgrading Dagger Engine in $module from $CURRENT_DAGGER_VERSION to $LATEST_DAGGER_VERSION"
    dagger -m "$module" develop
  else
    echo "$module is already using the latest Dagger Engine ($CURRENT_DAGGER_VERSION)"
  fi

}

#update dagger engine to latest version in all modules
upgrade_dagger_engine_all() {

  #upgrade dagger engine locally first
  brew upgrade dagger

  #create branch for updates
  LATEST_DAGGER_VERSION=$(detect_latest_dagger_version)
  git checkout -b "update_dagger_engine_$LATEST_DAGGER_VERSION"

  #upgrade dagger engine in modules
  for module in $(list_modules); do
    upgrade_dagger_engine "$module"
  done

  #upgrade dagger engine in test modules
  for module in $(list_modules_with_tests); do
    upgrade_dagger_engine "$module"
  done

  #git add and create commits for each module changed
  for module in $(list_modules); do
    # Get changed files under this module
    changed_files=$(git diff --name-only -- "$module")

    if [[ -n "$changed_files" ]]; then
      echo "📦 Module '$module' has changes:"
      echo "$changed_files"

      # Stage all changed files under the module
      echo "$changed_files" | xargs git add

      # Commit
      echo "Creating commit: fix($module): updating dagger engine to $LATEST_DAGGER_VERSION"
      git commit -S -m "fix($module): updating dagger engine to $LATEST_DAGGER_VERSION"
    else
      echo "No changes in $module"
    fi
  done


}

#find any act3 module updates and update release module with latest
upgrade_act3_module_deps() {

  dagger -m github.com/act3-ai/dagger/renovate call \
    --platform=github \
    --endpoint-url=https://api.github.com \
    --project=act3-ai/dagger \
    --author="$GITHUB_USER" \
    --email="$GITHUB_EMAIL" \
    --token=env:GITHUB_TOKEN \
    --git-private-key=env:GITHUB_PRIVATE_KEY \
    --enabled-managers="custom.regex" \
      update

}

#call function
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  if declare -f "$1" > /dev/null; then
    "$@"
  else
    echo "Error: '$1' is not a valid function."
    echo "Available functions:"
    declare -F | awk '{print $3}'
    exit 1
  fi
fi