# One-server-loss evidence contracts

These provider-neutral contracts define the private request, atomic fault
barrier, pinned data-probe protocol, and source-safe unsigned receipt used by
`cloudring-resilience one-server-loss`.

The observer is read-only. It establishes a stable healthy baseline and writes
a unique `ready-for-fault` marker. A separately approved downstream procedure
may then stop or reboot exactly the bound server. The same observer must remain
running through the loss and recovery windows; restarting it invalidates the
drill.

The request contains private Kubernetes names, label selectors, and a safe
query reference and must not be committed. The receipt contains only safe IDs,
counts, booleans, timestamps, and SHA-256 bindings. Raw UIDs, names, labels,
addresses, credentials, query output, and tenant data are excluded. Generated
markers and receipts remain deployment-private and require a separate trusted
signature before they can support a release claim.

Runtime validation in `pkg/resilience/oneserverloss` is authoritative. JSON
Schema cannot by itself prove timeline continuity, the exact one-unavailable
loss envelope, stable Kubernetes UID, VM recovery SLO, adapter executable
identity, or unchanged business-data digest. Preserving quorum after two
simultaneous losses does not satisfy this one-server-loss contract.
