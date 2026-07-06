# UI Extension Authoring

CloudRING portal extensions are OCSv3 (Open Cloud Standard 3) module
metadata plus a microfrontend host contract. The shell owns identity, tenant
context, navigation placement, permissions, and audit context. The service team
owns the module UI bundle and API contract.

## Required Surfaces

Declare:

- `service.spec.portalModules` with slot, route, API ref, host ref, mount ref,
  and permissions;
- `service.spec.ui.embedRef` and `contextSchemaRef`;
- host authority such as navigation, identity, and policy;
- allowed extension actions;
- `service.spec.ui.moduleHost` with host, runtime, mount ref, version range,
  integrity ref, sandbox, allowed events, and required context;
- UI certification evidence.

## Authoring Flow

1. Design the module around the shell context: tenant, project, subject, locale,
   theme, and permissions.
2. Publish a stable route and mount ref in the connector package.
3. Emit analytics events through declared event names.
4. Keep service API calls behind the module's public API refs.
5. Validate the connector package:

```sh
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
```

## Boundary Rules

The shell must not import service frontend internals, and the module must not
import platform internals. The contract is metadata, host context, API refs,
events, permissions, and evidence. This document does not claim production
readiness for any concrete UI bundle.
