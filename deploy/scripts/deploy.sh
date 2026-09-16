#!/usr/bin/env bash
set -euo pipefail

version="${1:-}"
if [[ ! "$version" =~ ^[0-9a-f]{40}$ ]]; then
  echo "usage: deploy.sh <full-git-sha>" >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
compose_file="$repo_root/deploy/compose/compose.prod.yml"
deploy_env="${PCP_DEPLOY_ENV:-/www/wwwroot/pcplatform/.env}"
state_dir="${PCP_DEPLOY_STATE_DIR:-/home/actions-runner/pcplatform-deploy}"
version_file="$state_dir/current-version"
backup_dir="$state_dir/backups"

if [[ ! -r "$deploy_env" ]]; then
  echo "deployment environment is not readable: $deploy_env" >&2
  exit 1
fi

mkdir -p "$backup_dir"
exec 9>"$state_dir/deploy.lock"
if ! flock -n 9; then
  echo "another production deployment is running" >&2
  exit 1
fi

previous_version=""
if [[ -f "$version_file" ]]; then
  previous_version="$(tr -d '[:space:]' < "$version_file")"
fi

export APP_VERSION="$version"
export PCP_ENV_FILE="$deploy_env"
compose=(docker compose --env-file "$deploy_env" -f "$compose_file")

"${compose[@]}" up -d postgres
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
"${compose[@]}" exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "$backup_dir/pcplatform-$timestamp.dump"

"${compose[@]}" pull migrate api worker public-web studio-web
"${compose[@]}" run --rm migrate
"${compose[@]}" up -d --no-build api worker public-web studio-web

healthy=true
for _ in $(seq 1 18); do
  if curl -fsS -o /dev/null "http://127.0.0.1:${PUBLIC_WEB_PORT:-23000}/" &&
     curl -fsS -o /dev/null "http://127.0.0.1:${STUDIO_WEB_PORT:-23001}/api/healthz"; then
    healthy=true
    break
  fi
  healthy=false
  sleep 5
done

if [[ "$healthy" != true ]]; then
  echo "health check failed for $version" >&2
  "${compose[@]}" ps >&2 || true
  if [[ "$previous_version" =~ ^[0-9a-f]{40}$ ]]; then
    echo "restoring application images from $previous_version" >&2
    export APP_VERSION="$previous_version"
    "${compose[@]}" up -d --no-build api worker public-web studio-web
  fi
  exit 1
fi

printf '%s\n' "$version" > "$version_file"
"${compose[@]}" ps
echo "production deployment completed: $version"
