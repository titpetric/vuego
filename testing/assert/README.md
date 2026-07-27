# Assert embeddings

This package exists to drop the dependency on stretchr/testify/assert.
It mutes the conversation on which to use (assert or require), and
provides a limited API that does about the same.

It's based on [fortio.org/assert](https://github.com/fortio/assert) and
contains the following modifications:

- Use `testing.TB` instead of explicit `*testing.T`
- Use `any` or `...any` for printf arguments

It's meant to be used from inside the current module and not
intended as an imported dependency.