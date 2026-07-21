#!/usr/bin/env bash


update_dag_deps() {
    local dagger_version="$1"

    # An associative array with, ie dag_deps[<name>]=<source>
    local -A dag_deps

    # Loop through all immediate subdirectories
    for dir in */; do
        # Remove trailing slash
        mod="${dir%/}"
        dagger_json_file="$mod/dagger.json"

        # folders that we want to skip
        if [[ "$mod" =~ ^(bin|\.dagger)$ ]]; then
            continue
        fi    

        # dagger.json file is required
        if [[ ! -f "$dagger_json_file" ]]; then
            printf "ERROR: dagger.json file for module: %s not found\n" "$mod"
            exit 1
        fi

        printf "\nDagger file: %s\n###########################\n\n" "$dagger_json_file"

        # Use jq to extract all dependencies <name>\t<source> 
        # Note: only dependencies with source that starts with "github.com/dagger/dagger"
        # Read name and source line-by-line using tab as a delimiter
        while IFS=$'\t' read -r name source; do
            if [[ -n "$name" ]]; then
                dag_deps["$name"]="$source"
            fi
        done < <(jq -r '.dependencies[]? | select(.source | startswith("github.com/dagger/dagger")) | "\(.name)\t\(.source)"' "$dagger_json_file")

        # Iterate over all loaded dependencies
        # DEVTODO for each you need to update the dep ex: dagger update wolfi@v0.21.6
        for name in "${!dag_deps[@]}"; do
            echo "Name: $name | Source: ${dag_deps[$name]}"
        done        
    done
}

printf "Go\n-----------------------------------------------------------\n"
update_dag_deps "v1.2.3"
printf "  \n-----------------------------------------------------------\nDone.\n"

