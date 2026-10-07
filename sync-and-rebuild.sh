#!/usr/bin/env bash

set -eo pipefail

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
    docker build -t owlenz/go-webapp:latest .
    minikube image load owlenz/go-webapp:latest
    kubectl apply -f k8s/ -R
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
  case "$1" in
  --sync-only)
    sync_source
    ;;
  *)
    sync_source
    build_and_restart_pod
    ;;
  esac
}

main "$@"
