#!/usr/bin/env bash


update_dag_deps() {
    local dagger_version="$1"

    # Loop through all immediate subdirs and any tests dir they may contain
    for dir in */ */tests/; do    
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

        # Use jq to extract all dependencies <name>\t<source> 
        # Note: only update dependencies that starts with "github.com/dagger/dagger"
        # Read name and source line-by-line using tab as a delimiter
        while IFS=$'\t' read -r name source; do
            if [[ -n "$name" ]]; then
                printf "Updating dependency: %s in %s to %s\n" "$name" "$mod" "$dagger_version"
                set -x
                dagger update -m "$mod" "${name}@${dagger_version}"
                set +x
            fi

        done < <(jq -r '.dependencies[]? | select(.source | startswith("github.com/dagger/dagger")) | "\(.name)\t\(.source)"' "$dagger_json_file")
    done
}

printf "Go\n-----------------------------------------------------------\n"
update_dag_deps "v0.21.6"
printf "  \n-----------------------------------------------------------\nDone.\n"

