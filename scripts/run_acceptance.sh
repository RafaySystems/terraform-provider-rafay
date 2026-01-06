#!/usr/bin/env bash

set -euo pipefail

export RCTL_PROJECT=${RCTL_PROJECT:-}
export RCTL_PROFILE=${RCTL_PROFILE:-}
export RCTL_REST_ENDPOINT=${RCTL_REST_ENDPOINT:-}
export RCTL_OPS_ENDPOINT=${RCTL_OPS_ENDPOINT:-}
export RCTL_API_KEY=${RCTL_API_KEY:-}
export RCTL_API_SECRET=${RCTL_API_SECRET:-}
export RCTL_SKIP_SERVER_CERT_VALIDATION=${RCTL_SKIP_SERVER_CERT_VALIDATION:-}
export RCTL_ORGANIZATION=${RCTL_ORGANIZATION:-}

# Delegate to go test so callers can add flags if needed.
TF_ACC=1 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay -run ^TestResourceBlueprintAcceptance$ -count=1
