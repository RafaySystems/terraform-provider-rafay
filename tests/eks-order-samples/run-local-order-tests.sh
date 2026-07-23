#!/usr/bin/env bash
set -eo pipefail

ROOT="/Users/vishalv/terraform-changes/terraform-provider-rafay"
STORE="$ROOT/.eks-local-backend.json"
export GOFLAGS="-mod=mod"
export GOCACHE="/tmp/go-build"
export GOMODCACHE="/tmp/go-modcache"
export GOTOOLCHAIN=auto
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

HARNESS=(go run "$ROOT/cmd/eks-local-backend")
CHECKER=(go run "$ROOT/tests/eks-order-samples/eks-order-check.go")

rm -f "$STORE"

write_yaml() {
  local path="$1"
  local managed_order="$2"
  local node_order="$3"
  cat > "$path" <<YAML
kind: Cluster
metadata:
  name: c1
  project: p1
spec:
  cloud_provider: dummy
---
kind: ClusterConfig
apiversion: rafay.io/v1alpha5
metadata:
  name: c1
  region: us-west-2
managedNodeGroups:
$managed_order
nodeGroups:
$node_order
YAML
}

mk_list() {
  if [ "$#" -eq 0 ]; then
    printf ""
    return 0
  fi
  local out=""
  for n in "$@"; do
    out+="- name: ${n}\n"
  done
  printf "%b" "$out"
}

mk_list_with_field() {
  local field="$1"
  local value="$2"
  shift 2
  if [ "$#" -eq 0 ]; then
    printf ""
    return 0
  fi
  local out=""
  for n in "$@"; do
    out+="- name: ${n}\n  ${field}: ${value}\n"
  done
  printf "%b" "$out"
}

mk_managed_taints_ng() {
  cat <<YAML
- name: taints-ng
  taints:
  - key: b
    effect: NoSchedule
    value: two
  - key: a
    effect: NoExecute
    value: one
YAML
}

mk_node_taints_ng() {
  cat <<YAML
- name: taints-ng
  taints:
  - key: b
    effect: NoSchedule
    value: two
  - key: a
    effect: NoExecute
    value: one
YAML
}

mk_managed_attach_ng() {
  cat <<YAML
- name: attach-ng
  securityGroups:
    attachIDs:
    - sg-3
    - sg-1
    - sg-2
YAML
}

mk_node_attach_ng() {
  cat <<YAML
- name: attach-ng
  securityGroups:
    attachIDs:
    - sg-3
    - sg-1
    - sg-2
YAML
}

mk_managed_instance_ng() {
  cat <<YAML
- name: instance-ng
  instanceTypes:
  - m6a.large
  - m5a.large
  - m7i.large
YAML
}

mk_node_instance_ng() {
  cat <<YAML
- name: instance-ng
  instancesDistribution:
    instanceTypes:
    - m6a.large
    - m5a.large
    - m7i.large
YAML
}

mk_managed_suspend_ng() {
  cat <<YAML
- name: suspend-ng
  asgSuspendProcesses:
  - Launch
  - AZRebalance
  - Terminate
YAML
}

mk_node_suspend_ng() {
  cat <<YAML
- name: suspend-ng
  asgSuspendProcesses:
  - Launch
  - AZRebalance
  - Terminate
YAML
}

mk_node_lb_ng() {
  cat <<YAML
- name: lb-ng
  classicLoadBalancerNames:
  - lb-3
  - lb-1
  - lb-2
YAML
}

mk_node_tg_ng() {
  cat <<YAML
- name: tg-ng
  targetGroupARNs:
  - tg-3
  - tg-1
  - tg-2
YAML
}

sort_names() {
  printf "%s\n" "$@" | sort | paste -sd, -
}

sorted_csv() {
  if [ -z "$1" ]; then
    printf ""
    return 0
  fi
  printf "%s\n" "$1" | tr ',' '\n' | sort | paste -sd, -
}

check_sorted() {
  local label="$1"
  local actual="$2"
  if [ -z "$actual" ]; then
    return 0
  fi
  local expected
  expected=$(sorted_csv "$actual")
  if [ "$expected" = "$actual" ]; then
    printf "%s: PASS\n" "$label"
  else
    printf "%s: FAIL (expected %s)\n" "$label" "$expected"
  fi
}

run_case() {
  local label="$1"
  local managed_names="$2"
  local node_names="$3"
  local managed_yaml="$4"
  local node_yaml="$5"
  local yaml_path="/tmp/${label}.yaml"

  write_yaml "$yaml_path" "$managed_yaml" "$node_yaml"
  "${HARNESS[@]}" apply --file "$yaml_path" > /dev/null
  "${HARNESS[@]}" read --name c1 --project p1 > "/tmp/${label}-read.yaml"

  local actual
  actual=$("${CHECKER[@]}" "/tmp/${label}-read.yaml")
  local managed_actual
  local node_actual
  managed_actual=$(echo "$actual" | awk -F= '/^managed=/{print $2}')
  node_actual=$(echo "$actual" | awk -F= '/^nodes=/{print $2}')

  local managed_expected
  local node_expected
  managed_expected=$(sort_names $managed_names)
  node_expected=$(sort_names $node_names)

  printf "\n== %s ==\n" "$label"
  printf "managed expected: %s\n" "$managed_expected"
  printf "managed actual  : %s\n" "$managed_actual"
  if [ "$managed_expected" = "$managed_actual" ]; then
    printf "managed result  : PASS\n"
  else
    printf "managed result  : FAIL\n"
  fi
  printf "nodes expected  : %s\n" "$node_expected"
  printf "nodes actual    : %s\n" "$node_actual"
  if [ "$node_expected" = "$node_actual" ]; then
    printf "nodes result    : PASS\n"
  else
    printf "nodes result    : FAIL\n"
  fi

  if echo "$actual" | grep -q '^managed.taints-ng='; then
    local taints
    taints=$(echo "$actual" | awk -F= '/^managed\.taints-ng=/{print $2}')
    printf "managed taints  : %s\n" "$taints"
    check_sorted "managed taints order" "$taints"
  fi
  if echo "$actual" | grep -q '^managed.attach-ng='; then
    local attach
    attach=$(echo "$actual" | awk -F= '/^managed\.attach-ng=/{print $2}')
    printf "managed attach  : %s\n" "$attach"
    check_sorted "managed attach_ids order" "$attach"
  fi
  if echo "$actual" | grep -q '^managed.instance-ng='; then
    local inst
    inst=$(echo "$actual" | awk -F= '/^managed\.instance-ng=/{print $2}')
    printf "managed inst    : %s\n" "$inst"
    check_sorted "managed instance_types order" "$inst"
  fi
  if echo "$actual" | grep -q '^managed.suspend-ng='; then
    local suspend
    suspend=$(echo "$actual" | awk -F= '/^managed\.suspend-ng=/{print $2}')
    printf "managed suspend : %s\n" "$suspend"
    check_sorted "managed asg_suspend_processes order" "$suspend"
  fi
  if echo "$actual" | grep -q '^node.taints-ng='; then
    local taintsn
    taintsn=$(echo "$actual" | awk -F= '/^node\.taints-ng=/{print $2}')
    printf "node taints     : %s\n" "$taintsn"
    check_sorted "node taints order" "$taintsn"
  fi
  if echo "$actual" | grep -q '^node.attach-ng='; then
    local attachn
    attachn=$(echo "$actual" | awk -F= '/^node\.attach-ng=/{print $2}')
    printf "node attach     : %s\n" "$attachn"
    check_sorted "node attach_ids order" "$attachn"
  fi
  if echo "$actual" | grep -q '^node.instance-ng='; then
    local instn
    instn=$(echo "$actual" | awk -F= '/^node\.instance-ng=/{print $2}')
    printf "node inst       : %s\n" "$instn"
    check_sorted "node instance_types order" "$instn"
  fi
  if echo "$actual" | grep -q '^node.suspend-ng='; then
    local suspendn
    suspendn=$(echo "$actual" | awk -F= '/^node\.suspend-ng=/{print $2}')
    printf "node suspend    : %s\n" "$suspendn"
    check_sorted "node asg_suspend_processes order" "$suspendn"
  fi
  if echo "$actual" | grep -q '^node.lb-ng='; then
    local lbs
    lbs=$(echo "$actual" | awk -F= '/^node\.lb-ng=/{print $2}')
    printf "node lbs        : %s\n" "$lbs"
    check_sorted "node classic_load_balancer_names order" "$lbs"
  fi
  if echo "$actual" | grep -q '^node.tg-ng='; then
    local tgs
    tgs=$(echo "$actual" | awk -F= '/^node\.tg-ng=/{print $2}')
    printf "node tgs        : %s\n" "$tgs"
    check_sorted "node target_group_arns order" "$tgs"
  fi
}

BASE=(vishal-001 vishal-002 vishal-003 shetty-001 shetty-002)

run_case "managed-start" "shetty-004 ${BASE[*]}" "" "$(mk_list shetty-004 "${BASE[@]}")" "$(mk_list)"
run_case "managed-middle" "vishal-001 vishal-002 shetty-004 vishal-003 shetty-001 shetty-002" "" "$(mk_list vishal-001 vishal-002 shetty-004 vishal-003 shetty-001 shetty-002)" "$(mk_list)"
run_case "managed-end" "${BASE[*]} shetty-004" "" "$(mk_list "${BASE[@]}" shetty-004)" "$(mk_list)"

run_case "node-start" "" "shetty-004 ${BASE[*]}" "$(mk_list)" "$(mk_list shetty-004 "${BASE[@]}")"
run_case "node-middle" "" "vishal-001 vishal-002 shetty-004 vishal-003 shetty-001 shetty-002" "$(mk_list)" "$(mk_list vishal-001 vishal-002 shetty-004 vishal-003 shetty-001 shetty-002)"
run_case "node-end" "" "${BASE[*]} shetty-004" "$(mk_list)" "$(mk_list "${BASE[@]}" shetty-004)"

run_case "unordered-backend" "shetty-002 vishal-003 vishal-001 shetty-001 vishal-002" "shetty-002 vishal-003 vishal-001 shetty-001 vishal-002" "$(mk_list shetty-002 vishal-003 vishal-001 shetty-001 vishal-002)" "$(mk_list shetty-002 vishal-003 vishal-001 shetty-001 vishal-002)"

run_case "backward-compat" "${BASE[*]}" "${BASE[*]}" "$(mk_list "${BASE[@]}")" "$(mk_list "${BASE[@]}")"

run_case "field-change" "${BASE[*]}" "${BASE[*]}" "$(mk_list_with_field instanceType t3.medium "${BASE[@]}")" "$(mk_list_with_field instanceType t3.medium "${BASE[@]}")"

run_case "managed-delete" "vishal-001 vishal-003 shetty-002" "" "$(mk_list vishal-003 shetty-002 vishal-001)" "$(mk_list)"

run_case "node-delete" "" "vishal-002 shetty-001 vishal-003" "$(mk_list)" "$(mk_list shetty-001 vishal-003 vishal-002)"

run_case "reorder-and-change" "shetty-002 vishal-003 vishal-001" "shetty-002 vishal-001 vishal-003" "$(mk_list_with_field instanceType c6i.large shetty-002 vishal-003 vishal-001)" "$(mk_list_with_field instanceType c6i.large shetty-002 vishal-001 vishal-003)"

run_case "taints-reorder" "taints-ng" "taints-ng" "$(mk_managed_taints_ng)" "$(mk_node_taints_ng)"

run_case "attach-ids-reorder" "attach-ng" "attach-ng" "$(mk_managed_attach_ng)" "$(mk_node_attach_ng)"

run_case "instance-types-reorder" "instance-ng" "instance-ng" "$(mk_managed_instance_ng)" "$(mk_node_instance_ng)"

run_case "suspend-processes-reorder" "suspend-ng" "suspend-ng" "$(mk_managed_suspend_ng)" "$(mk_node_suspend_ng)"

run_case "classic-lb-reorder" "" "lb-ng" "$(mk_list)" "$(mk_node_lb_ng)"

run_case "target-group-reorder" "" "tg-ng" "$(mk_list)" "$(mk_node_tg_ng)"
