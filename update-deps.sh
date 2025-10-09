#!/usr/bin/env bash

list_modules() {
find . -maxdepth 2 -mindepth 2 -type f -name dagger.json -exec dirname {} \; | sed 's|^\./||'
}

list_modules_with_tests() {
find . -maxdepth 3 -type f -path './*/tests/dagger.json' -exec dirname {} \; | sed 's|^\./||'
}

#update dagger engine to latest version in modules
upgrade_dagger_engine() {
brew upgrade dagger
#upgrade dagger engine in modules
  for module in $(list_modules); do
    echo "Upgrading Dagger Engine in $module"
    dagger -m "$module" develop
  done
#upgrade dagger engine in test modules
    for module in $(list_modules_with_tests); do
    echo "Upgrading Dagger Engine in $module"
    dagger -m "$module" develop
  done

}

upgrade_dagger_engine