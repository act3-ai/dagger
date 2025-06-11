#!/usr/bin/env bash
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
    version=$(dagger -m ../../github/dagger/git-cliff call --src="." bumped-version --args="--include-path=$module/**")
    #needed because version tag is format of module/v1.0.0
    stripped_version="${version#*/}"
    #generate and export new version/release notes
    dagger -m ../../github/dagger/release call --src="." prepare \
    --changelog "$module/CHANGELOG.md" \
    --notes-path "$module/releases/$stripped_version.md" \
    --ignore-error=$force \
    --args="--include-path=./$module/**" \
    export --path="."

    echo "Please review the local changes, especially $module/releases/$version.md"
    ;;

approve)
    version=$(dagger -m ../../github/dagger/git-cliff call --src="." bumped-version --args="--include-path=$module/**")
    #needed because version tag is format of module/v1.0.0
    stripped_version="${version#*/}"

    notesPath="$module/releases/$stripped_version.md"
    # release material
    git add "VERSION" "$module/CHANGELOG.md" "$notesPath"
    # documentation changes (from make gendoc, apidoc, swagger)
    # git add \*.md # updated
    # signed commit
    git commit -S -m "chore(release): prepare for $version"
    # annotated and signed tag
    git tag -s -a -m "Official release $version" "$version"
    ;;
publish)
    # push this branch and the associated tags
    git push --follow-tags

    version=$(dagger -m ../../github/dagger/git-cliff call --src="." bumped-version --args="--include-path=$module/**")
    notesPath="$module/releases/$stripped_version.md"
    
    # create release, upload artifacts
    dagger -m ../../github/dagger/release --src=. call \
        with-registry-auth --address="$registry" --username="$GITHUB_REG_USER" --secret=env://GITHUB_TOKEN  \
        create-github \
        --token=env://GITHUB_TOKEN \
        --host="github.com" \
        --project="act3-ai/dagger" \
        --version="$version" \
        --notes="$notesPath"

    ;;

*)
    help
    ;;
esac
