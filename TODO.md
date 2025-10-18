# TODO

## Completed

- [x] Update acceptance tests for new disk schema (2025-10-18)

## Notes

- Domain-level disk `<backingStore>` input is intentionally not supported; configure copy-on-write via the `libvirt_volume.backing_store` block until libvirt exposes the `backingStoreInput` feature to guests.
