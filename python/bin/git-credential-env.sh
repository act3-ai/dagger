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

[ -z "${host:-}" ] && exit 0

username_env="GIT_SECRET_USERNAME_${host}"
password_env="GIT_SECRET_PASSWORD_${host}"

username="${!username_env:-}"
password="${!password_env:-}"

[ -z "$username" ] && exit 0
[ -z "$password" ] && exit 0

echo "username=${username}"
echo "password=${password}"