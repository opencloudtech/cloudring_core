Kueue and scheduler ownership is modeled as an optional OCSv3 module package.

The manifest declares queue admission, fair sharing, optional Volcano
integration, compatibility windows, billing meters, and first-class
`not-installed` / `disabled` states. It is not a foundation prerequisite.

Validate the manifest with:

```sh
go run ./cmd/ocsctl validate ./cloudring_core/modules/scheduler/module-package.json
```
