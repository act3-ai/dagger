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

if [[ -z "$module" ]]; then
  echo "Module argument is required"
  exit 1
fi

case "$cmd" in
prepare)
    git fetch --tags

    #upgrade dagger engine
    echo "Upgrading dagger engine if needed.."
    upgrade_dagger_engine_and_commit "$module"
    upgrade_dagger_engine_and_commit "$module/tests"

    #run module tests
    dagger -m "$module/tests" checks

    dagger call --module="$module" prepare
    version=$(cat "$module/VERSION")

    echo "Please review the local changes, especially $module/releases/$version.md"
    if confirm_continue approve; then
      "$0" approve "$module"
    fi

    ;;

approve)
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
    # push this branch and the associated tags
    git push --follow-tags

    version=$(cat "$module/VERSION")

    dagger call --module="$module" release --version=$version

    ;;

*)
    help
    ;;
esac
