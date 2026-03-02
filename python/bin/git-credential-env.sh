#!/usr/bin/env bash
set -euo pipefail

action="$1"
[ "$action" = "get" ] || exit 0

unset host

while IFS='=' read -r key value || [ -n "$key" ]; do
	case "$key" in
		host) host="$value" ;;
	esac
done

# exit normally if host not provided, otherwise use creds
[ -z "${host:-}" ] && exit 0
# convert host to all uppercase, replacing dots with slashes so var expansion works
converted_host=$(echo "$host" | tr '[:lower:]' '[:upper:]' | tr '.' '_')
username_env="GIT_SECRET_USERNAME_${converted_host}"
password_env="GIT_SECRET_PASSWORD_${converted_host}"

username="${!username_env:-}"
password="${!password_env:-}"

[ -z "$username" ] && exit 0
[ -z "$password" ] && exit 0

echo "username=${username}"
echo "password=${password}"