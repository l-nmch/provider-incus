#!/usr/bin/env bash
# Dumps a normalized, secret-free snapshot of an Incus cluster's state as JSON
# on stdout, so that two snapshots taken before and after an e2e run can be
# diffed to prove the cluster was left untouched.
#
# Usage: hack/e2e/snapshot.sh [ssh-target] > snapshot.json
# Server config values are replaced with their sha256 so that secrets never
# land on disk while value changes remain detectable.
set -euo pipefail

TARGET="${1:-root@comp-opti-01.phorge}"

ssh -o BatchMode=yes "${TARGET}" python3 - <<'EOF'
import hashlib, json, subprocess

def q(path):
    out = subprocess.run(["incus", "query", path], capture_output=True, text=True)
    if out.returncode != 0:
        return {"error": out.stderr.strip()}
    return json.loads(out.stdout) if out.stdout.strip() else None

VOLATILE = ("last_used_at", "last_seen", "status", "status_code", "state",
            "location_status", "statusCode")

def clean(o):
    if isinstance(o, dict):
        return {k: clean(v) for k, v in sorted(o.items()) if k not in VOLATILE}
    if isinstance(o, list):
        items = [clean(v) for v in o]
        return sorted(items, key=lambda x: json.dumps(x, sort_keys=True))
    return o

snap = {}
server = q("/1.0")
snap["server_config"] = {k: hashlib.sha256(str(v).encode()).hexdigest()
                         for k, v in sorted(server["config"].items())}
snap["projects"] = q("/1.0/projects?recursion=1")
for ep in ("instances", "images", "profiles", "networks", "network-acls",
           "network-zones", "network-address-sets"):
    snap[ep] = q(f"/1.0/{ep}?recursion=1&all-projects=true")
snap["network-integrations"] = q("/1.0/network-integrations?recursion=1")
snap["storage-pools"] = q("/1.0/storage-pools?recursion=1")
for pool in snap["storage-pools"]:
    name = pool["name"]
    snap[f"storage-pools/{name}/volumes"] = q(f"/1.0/storage-pools/{name}/volumes?recursion=1&all-projects=true")
    snap[f"storage-pools/{name}/buckets"] = q(f"/1.0/storage-pools/{name}/buckets?recursion=1&all-projects=true")
for net in snap["networks"]:
    if not net.get("managed"):
        continue
    base = f"/1.0/networks/{net['name']}"
    proj = net.get("project", "default")
    for sub in ("forwards", "load-balancers", "peers"):
        snap[f"networks/{proj}/{net['name']}/{sub}"] = q(f"{base}/{sub}?recursion=1&project={proj}")
snap["certificates"] = q("/1.0/certificates?recursion=1")
snap["cluster-groups"] = q("/1.0/cluster/groups?recursion=1")
snap["cluster-members"] = [m["server_name"] for m in q("/1.0/cluster/members?recursion=1")]

print(json.dumps(clean(snap), indent=1, sort_keys=True))
EOF
