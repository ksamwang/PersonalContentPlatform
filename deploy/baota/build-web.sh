#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
npm ci
npm run build:web

mkdir -p apps/studio-web/.next/standalone/apps/studio-web/.next
mkdir -p apps/public-web/.next/standalone/apps/public-web/.next
cp -a apps/studio-web/.next/static apps/studio-web/.next/standalone/apps/studio-web/.next/
cp -a apps/public-web/.next/static apps/public-web/.next/standalone/apps/public-web/.next/

echo "Studio Web and Public Web production builds are ready."
