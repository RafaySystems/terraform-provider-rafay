#!/usr/bin/env bash
# Provide default RAFAY CI env vars (source before invoking tests)
: "${RCTL_PROJECT:=defaultproject}"
: "${RCTL_PROFILE:=dev}"
: "${RCTL_REST_ENDPOINT:=console-tb.dev.rafay-edge.net}"
: "${RCTL_OPS_ENDPOINT:=ops-console-tb.dev.rafay-edge.net}"
: "${RCTL_API_KEY:=**********}"
: "${RCTL_API_SECRET:=**********}"
: "${RCTL_SKIP_SERVER_CERT_VALIDATION:=false}"
: "${RCTL_ORGANIZATION:=defaultorg}"
: "${BASE_BLUEPRINT_VERSION:=4.0.0}"
export RCTL_PROJECT RCTL_PROFILE RCTL_REST_ENDPOINT RCTL_OPS_ENDPOINT \
  RCTL_API_KEY RCTL_API_SECRET RCTL_SKIP_SERVER_CERT_VALIDATION \
  RCTL_ORGANIZATION BASE_BLUEPRINT_VERSION
