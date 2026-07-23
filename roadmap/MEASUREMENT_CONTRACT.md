# Measurement contract

Claims about availability, zero downtime, scale, one-engineer operation and
developer simplicity are valid only when measured with a versioned profile.

## Common evidence envelope

Every measurement records:

- goal and requirement IDs;
- immutable release/artifact digests and deployed GitOps revision;
- sanitized hardware, topology and software-version profile;
- workload generator and dataset digests;
- start/end clocks in UTC and clock-sync health;
- warm-up, steady-state and recovery windows;
- request/operation counts, exclusions and formulas;
- raw evidence location and sanitized summary hash;
- abort thresholds, cleanup and verifier identity.

Changing a workload, topology, dataset or formula creates a new profile version;
results from different profiles are not silently compared.

## Availability and upgrade

- Monthly availability uses a complete consecutive calendar month and the
  published eligible-request definition. A prerelease soak cannot claim it.
- Goal qualification uses at least 24 continuous hours for the integrated
  management plane and at least the goal-specific failure/upgrade window.
- The 24-hour qualification passes only with at least 99.95% eligible-request
  availability, every profile-specific p99/error threshold green, zero committed
  data loss or invalid billing, and no unresolved release-blocking alert. It is a
  qualification result, not a claim that a calendar-month objective was observed.
- Availability denominator is all eligible synthetic and real test requests;
  excluded planned destructive tests are reported separately, never removed from
  a customer-facing SLO retrospectively.
- Zero-downtime upgrade means zero release-attributable failed eligible requests,
  zero unavailable readiness samples, zero lost accepted operations, zero data or
  billing corruption, and p99 remaining inside the published SLO. Probe interval
  is at most one second for the bounded upgrade campaign.
- Baseline latency/error distribution is measured for at least 30 minutes before
  mutation at the same load. The campaign reports absolute and relative change.

## Reference load profiles

G00 creates machine-readable profiles for at least:

- management API mix: reads, lists/watches and durable mutations by multiple
  tenants, including one noisy tenant;
- lifecycle mix: provision/resize/suspend/deprovision with provider latency and
  ambiguous failures;
- usage/billing stream: duplicates, late events, corrections and invoice close;
- portal/CLI/agent journey concurrency;
- each product's data/control path and failure-domain workload.

Each profile publishes p50/p95/p99 latency, throughput, saturation point, error
classes, queue depth, resource use and recovery time. A release has no generic
“high load” claim without these profiles.

Before G00 completes, the profiles also freeze the eligible-request definition,
minimum sample size, exact durable-operation acknowledgement threshold, exclusion
classes and pass formulas. Missing numeric thresholds block later performance or
“quick response” claims.

## Cell scale efficiency

At identical per-cell hardware and SLO, useful capacity means completed eligible
operations or documented product data units per steady-state second while every
latency, error, durability and fairness threshold remains green. The useful
capacity of two independent cells must be at least 1.7 times one-cell capacity
before G25 claims horizontal scale.
The report identifies the first shared bottleneck and reserves at least 30%
headroom at the recommended operating point. Tenant/system fairness and recovery
traffic must pass while one cell is deliberately overloaded.

## Installation and operator toil

- Installation timing starts after the versioned prerequisite validator reports
  green and ends after live acceptance. Waiting for hardware or approvals is
  reported separately.
- Human-attention time counts active terminal/UI, diagnosis and manual-decision
  time; unattended reconciliation does not count. Repeated failed automation does.
- The one-engineer profile is a trained Linux/Kubernetes operator using only
  public docs and shipped tooling, with no author/private chat assistance.
- Healthy-state daily toil is sampled for at least 14 representative days.
- The incident set includes API replica loss, database failover, GitOps drift,
  expired/rotating credential, connector outage, network/storage degradation,
  capacity risk and failed upgrade.

## Developer experience

- The tester is a mid-level developer familiar with one supported language but
  not CloudRING internals.
- Timing starts from a clean machine and empty repository after documented tool
  prerequisites are installed; it ends when the original service passes local
  positive/negative conformance and produces a verified signed package.
- Permitted help is public documentation and ordinary compiler/test errors. No
  private repository, unpublished module, author cache or direct author guidance
  is allowed.
- Report setup time, active development time, conformance failures, documentation
  defects and the exact package digest. The two-hour target is not met by a
  generated unchanged template.
