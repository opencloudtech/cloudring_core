# GitOps ownership evidence contract

CloudRING exposes a provider-neutral Go contract in `pkg/gitopsownership` for
proving that an exact accepted Flux source artifact owns every declared
critical resource through an exact set of Kustomization roots.

The contract binds:

- the accepted Git source revision and artifact digest;
- the public CloudRING gitlink revision used by the downstream repository;
- each selected root's source reference, path, dependency graph, suspension,
  pruning, wait, deletion-policy, and spec digest;
- each root's observed generation, readiness, applied revision, and inventory;
- the unique expected owner of every exact critical resource.

Missing roots, mutable or mismatched source evidence, stale readiness,
suspended roots, premature pruning, spec drift, duplicate ownership, incomplete
families, and unmanaged resources fail closed. Verification is read-only and
never enables prune or performs the controlled drift mutation.

Provider repositories own their live collectors and site contracts. A collector
must obtain the accepted revision and digest from the already accepted
downstream commit and Flux `GitRepository.status.artifact`, then populate the
public snapshot types without copying endpoints, credentials, tenant data, or
raw logs into evidence. Passing the library validator does not claim that a
collector is trustworthy, that a deployment is healthy, or that a drift drill
was executed; those are separate signed live-evidence gates.
