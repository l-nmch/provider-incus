#!/usr/bin/env bash
# Idempotently removes everything an e2e run may have created, on both the
# Kubernetes cluster and the Incus cluster. Safe to run at any point, including
# after a failed or interrupted run.
#
# Usage: hack/e2e/cleanup.sh [kube-context] [ssh-target]
# Every Incus object created by the e2e suite is prefixed with "xptest".
set -uo pipefail

CONTEXT="${1:-control}"
TARGET="${2:-root@comp-opti-01.phorge}"
NS=xptest
k() { kubectl --context "${CONTEXT}" "$@"; }

echo ">> Kubernetes (${CONTEXT})"
if k get ns "${NS}" >/dev/null 2>&1; then
  # Ask the provider (if still running) to delete external resources first.
  for crd in $(k get crd -o name | grep -E 'incus\.(m\.)?crossplane\.io' | grep -v providerconfig | sed 's|.*/||'); do
    k -n "${NS}" delete "${crd}" --all --wait=false >/dev/null 2>&1
  done
  sleep 5
  # Strip finalizers so nothing blocks on a stopped provider; leftovers on the
  # Incus side are handled below.
  for crd in $(k get crd -o name | grep -E 'incus\.(m\.)?crossplane\.io' | sed 's|.*/||'); do
    for o in $(k -n "${NS}" get "${crd}" -o name 2>/dev/null); do
      k -n "${NS}" patch "${o}" --type merge -p '{"metadata":{"finalizers":null}}' >/dev/null 2>&1
    done
  done
  k delete ns "${NS}" --wait=true --timeout=120s
fi
for crd in $(k get crd -o name | grep -E 'incus\.(m\.)?crossplane\.io'); do
  for o in $(k get "${crd#*/}" -o name 2>/dev/null); do
    k patch "${o}" --type merge -p '{"metadata":{"finalizers":null}}' >/dev/null 2>&1
    k delete "${o}" --wait=false >/dev/null 2>&1
  done
  k delete "${crd}" --wait=true --timeout=60s
done

echo ">> Incus (${TARGET})"
ssh -o BatchMode=yes "${TARGET}" bash -s <<'EOF'
set -u
projects=$(incus project list -f csv -c n | sed 's/ (.*)//' | grep '^xptest' || true)
for p in $projects default; do
  for i in $(incus list --project "$p" -f csv -c n | grep '^xptest'); do incus delete -f --project "$p" "$i"; done
  for f in $(incus image list --project "$p" -f csv -c f); do
    incus image show --project "$p" "$f" | grep -q 'xptest' && incus image delete --project "$p" "$f"
  done
  for pool in $(incus storage list -f csv -c n); do
    for b in $(incus storage bucket list "$pool" --project "$p" -f csv -c n 2>/dev/null | grep '^xptest'); do incus storage bucket delete "$pool" "$b" --project "$p"; done
    for v in $(incus storage volume list "$pool" --project "$p" -f csv -c tn type=custom 2>/dev/null | grep ',xptest' | cut -d, -f2); do incus storage volume delete "$pool" "$v" --project "$p"; done
  done
  for n in $(incus network list --project "$p" -f csv -c n | grep '^xptest'); do
    for a in $(incus network forward list "$n" --project "$p" -f csv -c l 2>/dev/null); do incus network forward delete "$n" "$a" --project "$p"; done
    for a in $(incus network load-balancer list "$n" --project "$p" -f csv -c l 2>/dev/null); do incus network load-balancer delete "$n" "$a" --project "$p"; done
    for a in $(incus network peer list "$n" --project "$p" -f csv -c n 2>/dev/null); do incus network peer delete "$n" "$a" --project "$p"; done
  done
  for n in $(incus network list --project "$p" -f csv -c n | grep '^xptest'); do incus network delete "$n" --project "$p"; done
  for z in $(incus network zone list --project "$p" -f csv -c n | grep '^xptest'); do
    for r in $(incus network zone record list "$z" --project "$p" -f csv -c n); do incus network zone record delete "$z" "$r" --project "$p"; done
    incus network zone delete "$z" --project "$p"
  done
  for a in $(incus network acl list --project "$p" -f csv | cut -d, -f1 | grep '^xptest'); do incus network acl delete "$a" --project "$p"; done
  for a in $(incus network address-set list --project "$p" -f csv 2>/dev/null | cut -d, -f1 | grep '^xptest'); do incus network address-set delete "$a" --project "$p"; done
  for pr in $(incus profile list --project "$p" -f csv -c n | grep '^xptest'); do incus profile delete "$pr" --project "$p"; done
done
for p in $projects; do incus project delete "$p" --force; done
for i in $(incus network integration list -f csv -c n 2>/dev/null | grep '^xptest'); do incus network integration delete "$i"; done
for pool in $(incus storage list -f csv -c n | grep '^xptest'); do incus storage delete "$pool"; done
for g in $(incus cluster group list -f csv -c n | grep '^xptest'); do incus cluster group delete "$g"; done
[ -n "$(incus config get user.xptest)" ] && incus config unset user.xptest
for fp in $(incus query /1.0/certificates?recursion=1 | python3 -c 'import json,sys;[print(c["fingerprint"]) for c in json.load(sys.stdin) if c["name"].startswith("xptest")]'); do
  incus config trust remove "$fp"
done
echo done
EOF
