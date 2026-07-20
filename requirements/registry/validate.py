#!/usr/bin/env python3
"""CloudRING requirements corpus validator and registry generator.

Stdlib-only. Parses every requirement block in ../domains/*.md against
../01-requirement-schema.md, regenerates requirements.json, and reports
schema violations, cross-reference gaps, and registry drift.

Usage:
    python3 validate.py            regenerate requirements.json and validate
    python3 validate.py --check    validate only; fail if requirements.json
                                   on disk differs from the corpus (CI drift
                                   gate; nothing is written)

Exit code is 0 when there are no errors (warnings are reported but do not
fail), 1 otherwise. Wording is English only by corpus policy.
"""

import json
import re
import sys
from pathlib import Path

REGISTRY_DIR = Path(__file__).resolve().parent
REQUIREMENTS_ROOT = REGISTRY_DIR.parent
DOMAINS_DIR = REQUIREMENTS_ROOT / "domains"
REGISTRY_JSON = REGISTRY_DIR / "requirements.json"

# --- Schema constants (mirrors 01-requirement-schema.md) --------------------

DOMAIN_FILES = {
    "FND": "10-platform-foundation.md",
    "CMP": "11-compute-virtualization.md",
    "NET": "12-network.md",
    "STO": "13-storage-backup-dr.md",
    "K8S": "14-kubernetes-containers.md",
    "IAM": "15-iam-identity-security.md",
    "BIL": "16-billing-finops.md",
    "OCS": "17-ocs-service-connectors.md",
    "MKT": "18-marketplace-catalog.md",
    "CUX": "19-portal-ux-selfservice.md",
    "OBS": "20-observability.md",
    "OPS": "21-ops-sre-support.md",
    "DPL": "22-deployment-iac-cicd.md",
    "FED": "23-federation-global-portal.md",
    "DAT": "24-data-services.md",
    "AGT": "25-agent-governance.md",
}

REQUIRED_FIELDS = [
    "Priority",
    "Status",
    "Actors",
    "Problem",
    "Requirement",
    "Acceptance evidence",
    "Non-goals",
    "Non-claims",
    "Stop conditions",
    "Traceability",
]

PRIORITIES = {"P0", "P1", "P2"}
STATUSES = {"proposed", "accepted", "blocked", "retired"}
ACTORS = {
    "provider",
    "vendor",
    "service-team",
    "tenant",
    "operator",
    "agent",
    "auditor",
}

PROVENANCE_CLASSES = [
    "vision-deck",
    "req-ccp",
    "req-acr-singular",
    "req-acr-plural",
    "req-history",
    "legacy-platform-a",
    "legacy-platform-b",
    "current-core",
]

RISK_CLASSES = ["money", "data", "keys", "trust", "exposure", "deletion",
                "migration", "settlement"]

# Reference-platform brands are banned anywhere in corpus text
# (02-source-and-method.md, rule 2). Open-source ecosystem building blocks
# named as implementation classes (Kubernetes, KubeVirt, etcd, ...) are not
# reference platforms and are not banned.
BANNED_BRANDS = [
    "vmware", "vsphere", "vcenter", "aws", "amazon web services", "azure",
    "gcp", "google cloud", "alibaba cloud", "aliyun", "yandex cloud",
    "selectel", "vk cloud", "openstack", "huawei cloud", "tencent cloud",
    "oracle cloud", "ibm cloud", "digitalocean", "hetzner", "ovh",
    "hyper-v", "proxmox",
]

# Specific third-party product identifiers are discouraged in corpus text;
# reported as warnings for owner review rather than hard errors.
PRODUCT_IDENTIFIERS = ["netbox"]

# Words that indicate a risk class when they appear in requirement text.
# Used to audit `n/a` stop conditions.
RISK_WORDS = {
    "money": ["money", "billing", "billed", "invoice", "payment", "charge",
              "charged", "cost", "price", "pricing", "refund", "arrears",
              "fee", "revenue"],
    "data": ["data loss", "data integrity", "data durability",
             "data deletion", "data corruption", "corrupt", "metadata"],
    "keys": ["key", "keys", "secret", "secrets", "credential", "credentials",
             "token", "tokens", "password"],
    "trust": ["trust", "trusted", "untrusted"],
    "exposure": ["exposure", "expose", "exposed", "exposing", "public"],
    "deletion": ["delete", "deletes", "deletion", "purge", "destroy",
                 "destroys", "reclaim"],
    "migration": ["migration", "migrate", "migrates", "migrating",
                  "renumbering"],
    "settlement": ["settlement", "settle", "settles", "reconciliation"],
}

# Non-ASCII characters accepted as English typographic punctuation.
ALLOWED_NON_ASCII = set("—–‘’“”…→↔⇒≤≠")

HEADER_RE = re.compile(r"^###\s+(CR-([A-Z0-9]{3})-(\d{3}))\s+—\s+(.+?)\s*$")
LOOSE_HEADER_RE = re.compile(r"^###\s*(CR-\S+)")
FIELD_RE = re.compile(r"^-\s+\*\*([^:*]+):\*\*\s?(.*)$")
ID_REF_RE = re.compile(r"\bCR-[A-Z0-9]{3}-\d{3}\b")


class Finding:
    def __init__(self, severity, req_id, file, message):
        self.severity = severity  # "error" | "warning"
        self.req_id = req_id
        self.file = file
        self.message = message

    def __str__(self):
        where = self.req_id or self.file or "corpus"
        return f"[{self.severity}] {where}: {self.message}"


def split_blocks(lines):
    """Yield (start_lineno, [block lines]) for each '### ' block in a file."""
    start = None
    block = []
    for lineno, line in enumerate(lines, 1):
        if line.startswith("### "):
            if block:
                yield start, block
            start = lineno
            block = [line]
        elif block:
            if line.strip() == "---":
                yield start, block
                start, block = None, []
            else:
                block.append(line)
    if block:
        yield start, block


def parse_fields(block_lines):
    """Parse '- **Field:** value' lines with continuation lines."""
    fields = {}
    current = None
    for line in block_lines[1:]:
        m = FIELD_RE.match(line)
        if m:
            current = m.group(1).strip()
            fields.setdefault(current, [])
            fields[current].append(m.group(2).strip())
        elif current is not None:
            fields[current].append(line.strip())
    return {k: " ".join(v).strip() for k, v in fields.items()}


def parse_actors(value):
    """Split an Actors value; return (actors, separator)."""
    if "|" in value:
        parts = value.split("|")
        sep = "|"
    else:
        parts = value.split(",")
        sep = ","
    return [p.strip() for p in parts if p.strip()], sep


def extract_risk_classes(stop_text):
    """Risk classes mentioned in a Stop conditions value."""
    found = []
    low = stop_text.lower()
    for cls in RISK_CLASSES:
        words = [cls]
        if cls == "keys":
            words.append("key")
        for w in words:
            if re.search(r"\b" + re.escape(w) + r"\b", low):
                found.append(cls)
                break
    return found


def extract_provenance(trace_text):
    """Known provenance classes mentioned in a Traceability value."""
    found = []
    for cls in PROVENANCE_CLASSES:
        if re.search(r"\b" + re.escape(cls) + r"\b", trace_text):
            found.append(cls)
    return found


def risk_words_in(text):
    """Risk classes whose indicator words appear in free text."""
    low = text.lower()
    hits = {}
    for cls, words in RISK_WORDS.items():
        matched = [w for w in words
                   if re.search(r"\b" + re.escape(w) + r"\b", low)]
        if matched:
            hits[cls] = matched
    return hits


def parse_corpus():
    """Parse all domain files. Return (requirements, findings, stats)."""
    findings = []
    requirements = []
    seen_ids = {}
    file_actor_sep = {}

    expected_files = {name for name in DOMAIN_FILES.values()}
    actual_files = {p.name for p in sorted(DOMAINS_DIR.glob("*.md"))}
    for name in sorted(expected_files - actual_files):
        findings.append(Finding("error", None, name,
                                "domain file listed in schema is missing"))
    for name in sorted(actual_files - expected_files):
        findings.append(Finding("error", None, name,
                                "file not mapped to a domain code in schema"))

    for path in sorted(DOMAINS_DIR.glob("*.md")):
        rel = f"domains/{path.name}"
        lines = path.read_text(encoding="utf-8").splitlines()
        for lineno, block_lines in split_blocks(lines):
            header = HEADER_RE.match(block_lines[0])
            if not header:
                loose = LOOSE_HEADER_RE.match(block_lines[0])
                rid = loose.group(1) if loose else block_lines[0][:40]
                findings.append(Finding(
                    "error", rid, rel,
                    f"line {lineno}: malformed requirement header; expected "
                    "'### CR-<DOMAIN>-NNN — <title>' (check ID spacing and "
                    "the em-dash separator)"))
                continue
            req_id, domain_code, number, title = header.groups()

            if domain_code not in DOMAIN_FILES:
                findings.append(Finding(
                    "error", req_id, rel,
                    f"unknown domain code '{domain_code}'"))
            elif DOMAIN_FILES[domain_code] != path.name:
                findings.append(Finding(
                    "error", req_id, rel,
                    f"ID domain code {domain_code} does not match file "
                    f"{path.name} (expected {DOMAIN_FILES[domain_code]})"))
            if int(number) % 10 != 0:
                findings.append(Finding(
                    "error", req_id, rel,
                    f"ID number {number} is not a multiple of 10 (wrong ID "
                    "spacing; numbering increments by 10)"))
            if req_id in seen_ids:
                findings.append(Finding(
                    "error", req_id, rel,
                    f"duplicate ID; first defined in "
                    f"{seen_ids[req_id]}"))
            seen_ids.setdefault(req_id, rel)

            fields = parse_fields(block_lines)
            for field in REQUIRED_FIELDS:
                if field not in fields:
                    findings.append(Finding(
                        "error", req_id, rel,
                        f"missing required field '{field}'"))
                elif not fields[field]:
                    findings.append(Finding(
                        "error", req_id, rel,
                        f"field '{field}' is empty"))

            priority = fields.get("Priority", "")
            if priority and priority not in PRIORITIES:
                findings.append(Finding(
                    "error", req_id, rel,
                    f"invalid priority '{priority}'"))

            status = fields.get("Status", "")
            if status and status not in STATUSES:
                findings.append(Finding(
                    "error", req_id, rel,
                    f"invalid status '{status}'"))

            actors, sep = parse_actors(fields.get("Actors", ""))
            file_actor_sep.setdefault(rel, set()).add(sep)
            if not actors:
                findings.append(Finding(
                    "error", req_id, rel, "no actors listed"))
            for actor in actors:
                if actor not in ACTORS:
                    findings.append(Finding(
                        "error", req_id, rel,
                        f"unknown actor '{actor}'"))

            block_text = "\n".join(block_lines)
            for ch in block_text:
                if ord(ch) > 127 and ch not in ALLOWED_NON_ASCII:
                    findings.append(Finding(
                        "error", req_id, rel,
                        f"non-English/non-ASCII character {ch!r} "
                        f"(U+{ord(ch):04X}) in block text"))
                    break
            low_text = block_text.lower()
            for brand in BANNED_BRANDS:
                if re.search(r"\b" + re.escape(brand) + r"\b", low_text):
                    findings.append(Finding(
                        "error", req_id, rel,
                        f"banned reference-platform brand name '{brand}'"))
            for product in PRODUCT_IDENTIFIERS:
                if re.search(r"\b" + re.escape(product) + r"\b", low_text):
                    findings.append(Finding(
                        "warning", req_id, rel,
                        f"specific third-party product identifier "
                        f"'{product}' in corpus text"))

            stop = fields.get("Stop conditions", "")
            risk_classes = extract_risk_classes(stop)
            if stop:
                normalized = stop.lstrip().lower()
                if normalized.startswith("n/a"):
                    body = " ".join(
                        fields.get(f, "")
                        for f in ("Problem", "Requirement", "Non-goals"))
                    hits = risk_words_in(body)
                    if hits:
                        detail = "; ".join(
                            f"{cls} ({', '.join(sorted(ws))})"
                            for cls, ws in sorted(hits.items()))
                        findings.append(Finding(
                            "warning", req_id, rel,
                            f"stop conditions are 'n/a' but risk-class "
                            f"words appear in requirement text: {detail}"))

            trace = fields.get("Traceability", "")
            provenance = extract_provenance(trace)
            if trace and not provenance:
                findings.append(Finding(
                    "error", req_id, rel,
                    "traceability names no known provenance class"))

            related = sorted({r for r in ID_REF_RE.findall(block_text)
                              if r != req_id})

            requirements.append({
                "id": req_id,
                "title": title,
                "domain": domain_code,
                "file": rel,
                "priority": priority,
                "status": status,
                "actors": actors,
                "risk_classes": risk_classes,
                "traceability": provenance,
                "_related": related,
            })

    # Wrong ID spacing within a file: out-of-order numbering.
    by_file = {}
    for req in requirements:
        by_file.setdefault(req["file"], []).append(req)
    for rel, reqs in sorted(by_file.items()):
        numbers = [int(r["id"].rsplit("-", 1)[1]) for r in reqs]
        if numbers != sorted(numbers):
            findings.append(Finding(
                "error", None, rel,
                "requirement IDs are not in ascending order within the "
                "file"))

    # Inconsistent Actors separator within one file.
    for rel, seps in sorted(file_actor_sep.items()):
        if len(seps) > 1:
            findings.append(Finding(
                "error", None, rel,
                f"mixed Actors separators in one file: "
                f"{', '.join(sorted(seps))}"))

    # Corpus-wide Actors separator uniformity.
    sep_by_file = {rel: next(iter(seps)) for rel, seps in
                   file_actor_sep.items() if len(seps) == 1}
    styles = {sep for sep in sep_by_file.values()}
    if len(styles) > 1:
        majority = max(styles, key=lambda s: list(
            sep_by_file.values()).count(s))
        for rel, sep in sorted(sep_by_file.items()):
            if sep != majority:
                findings.append(Finding(
                    "warning", None, rel,
                    f"Actors lists use '{sep}' as separator while the "
                    f"corpus majority uses '{majority}'; schema examples "
                    "enumerate allowed values, so one list separator "
                    "should be used corpus-wide"))

    # Cross-reference existence.
    known_ids = set(seen_ids)
    for req in requirements:
        for ref in req["_related"]:
            if ref not in known_ids:
                findings.append(Finding(
                    "error", req["id"], req["file"],
                    f"cross-referenced requirement {ref} does not exist"))

    return requirements, findings


def registry_records(requirements):
    """Strip internal keys and order records by schema file order, then ID."""
    order = {name: i for i, name in enumerate(DOMAIN_FILES.values())}
    sorted_reqs = sorted(
        requirements,
        key=lambda r: (order.get(r["file"].split("/", 1)[1], 99), r["id"]))
    return [
        {
            "id": r["id"],
            "title": r["title"],
            "domain": r["domain"],
            "file": r["file"],
            "priority": r["priority"],
            "status": r["status"],
            "actors": r["actors"],
            "risk_classes": r["risk_classes"],
            "traceability": r["traceability"],
        }
        for r in sorted_reqs
    ]


def main():
    check_only = "--check" in sys.argv[1:]
    requirements, findings = parse_corpus()
    records = registry_records(requirements)
    payload = json.dumps(records, indent=2, ensure_ascii=False) + "\n"

    if check_only:
        if not REGISTRY_JSON.exists():
            findings.append(Finding(
                "error", None, "registry/requirements.json",
                "registry is missing; run validate.py to generate it"))
        elif REGISTRY_JSON.read_text(encoding="utf-8") != payload:
            findings.append(Finding(
                "error", None, "registry/requirements.json",
                "registry drift: requirements.json differs from the "
                "Markdown corpus; regenerate with validate.py"))
    else:
        REGISTRY_JSON.write_text(payload, encoding="utf-8")

    errors = [f for f in findings if f.severity == "error"]
    warnings = [f for f in findings if f.severity == "warning"]

    by_domain = {}
    by_priority = {}
    by_status = {}
    for r in records:
        by_domain[r["domain"]] = by_domain.get(r["domain"], 0) + 1
        by_priority[r["priority"]] = by_priority.get(r["priority"], 0) + 1
        by_status[r["status"]] = by_status.get(r["status"], 0) + 1

    print(f"Parsed {len(records)} requirements from "
          f"{len(DOMAIN_FILES)} domain files.")
    print("Per domain: " + ", ".join(
        f"{k}={by_domain[k]}" for k in sorted(by_domain)))
    print("Per priority: " + ", ".join(
        f"{k}={by_priority[k]}" for k in sorted(by_priority)))
    print("Per status: " + ", ".join(
        f"{k}={by_status[k]}" for k in sorted(by_status)))

    for f in findings:
        print(f, file=sys.stderr)

    print(f"Validation: {len(errors)} error(s), "
          f"{len(warnings)} warning(s).")
    if check_only:
        print("Drift check: registry compared against corpus "
              "(nothing written).")
    else:
        print(f"Wrote {REGISTRY_JSON}")

    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
