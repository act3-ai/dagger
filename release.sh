#!/usr/bin/env bash

# Required env vars:
# GITHUB_TOKEN - github repo api access

force=false
cmd=$1
shift

module=""

# Loop through remaining args
while [[ $# -gt 0 ]]; do
  case "$1" in
    -f|--force)
      force=true
      shift
      ;;
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
    #run module tests
    #dagger -m "$module"/tests call all

    version=$(
      dagger -m release call --git-ref="." \
      version \
      --config="$module/cliff.toml"
    )
    
    #generate and export new version/release notes
    dagger -m release call --git-ref="." -v prepare \
    --path-prefix="$module" \
    --version="$version" \
    --token=env://GITHUB_TOKEN \
    export --path="."

    echo "Please review the local changes, especially $module/releases/$version.md"
    ;;

approve)
    version=$(cat "$module/VERSION")

    notesPath="$module/releases/v$version.md"
    # release material
    git add "$module/VERSION" "$module/CHANGELOG.md" "$notesPath"
    # documentation changes (from make gendoc, apidoc, swagger)
    # git add \*.md # updated
    # signed commit
    git commit -S -m "chore(release): prepare for $module/v$version"
    # annotated and signed tag
    git tag -s -a -m "Official release $module/v$version" "$module/v$version"
    ;;
publish)
    # push this branch and the associated tags
    git push --follow-tags

    version=$(cat "$module/VERSION")
    notesPath="$module/releases/v$version.md"
    
    # create release, upload artifacts
    dagger -m release --git-ref="." call \
        create-github \
        --token=env://GITHUB_TOKEN \
        --repo="act3-ai/dagger" \
        --title="$module/v$version" \
        --version="$module/v$version" \
        --notes="$notesPath"

    ;;

*)
    help
    ;;
esac
