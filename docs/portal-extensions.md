# Portal Extensions

Service UI is delivered as a module extension. The platform shell provides
navigation, identity context, billing/support links, and extension slots.

Each service module declares:

- portal module name, slot, route, API ref, host ref, mount ref, permissions;
- microfrontend host runtime and sandbox;
- integrity evidence;
- allowed events and required context.

Do not hardcode service-specific routes into platform core. The platform reads
the module package and mounts extensions dynamically.
