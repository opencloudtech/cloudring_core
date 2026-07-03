# Tutorial: Build A Service Module

1. Copy the synthetic reference service.
2. If your shell has no `cp`, copy the directory with the equivalent file
   manager or platform command.
3. Rename `metadata.name`, `catalog.serviceClass`, Kubernetes binding, and CRD.
4. Keep lifecycle actions complete.
5. Add service-specific diagnostics and evidence refs.
6. Run:

```bash
go run ./cmd/ocsctl validate ./my-service/module-package.json
go run ./cmd/ocsctl conformance ./my-service/module-package.json
```

If conformance fails, fix the reported field before adding platform wiring.
