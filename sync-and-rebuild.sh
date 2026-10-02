#!/usr/bin/env bash

set -euo pipefail

DIR_NAME="k8s_playground"
SRC="$(cd "$(dirname "$0")/.." && pwd)/$DIR_NAME"
REMOTE="owlenz@localhost"
REMOTE_DIR="/home/owlenz/dev"

function sync_source {
  rsync -avh --delete -e 'ssh -p 2222' "$SRC" "$REMOTE:$REMOTE_DIR"
}

function build_and_restart_pod {
  ssh -p 2222 "$REMOTE" "
    set -e
    export PATH=\"\$HOME/.nix-profile/bin:\$PATH\"
    cd $REMOTE_DIR/$DIR_NAME
    docker build -t go-webapp:v1 .
    minikube image load go-webapp:v1
    kubectl rollout restart deployment/webapp-deployment
    kubectl rollout status deployment/webapp-deployment --timeout=60s
  "
}

function die {
  echo "error: $*" >&2
  exit 1
}

main() {
  command -v rsync >/dev/null || die "rsync not installed"
  sync_source
  build_and_restart_pod
}

main "$@"
