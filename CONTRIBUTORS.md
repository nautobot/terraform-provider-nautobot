# Contributing to the Nautobot Terraform Provider

Thank you for contributing. These conventions keep the provider predictable for Terraform and
OpenTofu users and make changes easier to review and maintain.

Participation in this project is governed by the [Code of Conduct](.github/CODE_OF_CONDUCT.md).

## Before starting

- Search existing issues and pull requests before beginning substantial work.
- Open an issue before making a breaking change or introducing a large new dependency.
- Keep pull requests focused. Separate unrelated cleanup or refactoring from functional changes.
- Never run acceptance tests against a production Nautobot instance. Acceptance tests create,
  update, and delete real objects.

The required Go version is declared in [`go.mod`](go.mod). Acceptance-test Terraform and
OpenTofu versions are declared in [`.github/workflows/test.yml`](.github/workflows/test.yml).

## Development workflow

Create a branch, make the change, and run the relevant checks before opening a pull request:

```sh
make fmt
make fmt-check
make test
```

If a provider schema or its descriptions changed, regenerate and commit the documentation:

```sh
make docs
```

Do not hand-edit generated schema sections as a substitute for changing the Go schema. Review all
generated changes before committing them.

## Go conventions

- Format all Go files with `gofmt`; CI enforces `make fmt-check`.
- Prefer small, purpose-specific functions and shared helpers for genuinely repeated API mapping.
- Add direct unit tests for shared helpers, including unset, explicit-null, malformed, and populated
  values where applicable.
- Propagate Terraform framework diagnostics and stop processing after fatal diagnostics.
- Include enough context in API errors to identify the operation and object without exposing secrets.
- Preserve the request context supplied by Terraform instead of replacing it with a background
  context in provider implementation code.
- Avoid unrelated rewrites and formatting churn.

## Provider schema conventions

- Treat public resource and data-source schemas as APIs. Renaming or removing an attribute is a
  breaking change and requires an explicit migration plan.
- Name foreign-key UUID attributes with an `_id` suffix. Use the established provider spelling for
  existing collection attributes, such as `tags_ids`, unless a compatibility plan is agreed first.
- Make configurable attributes `Required` or `Optional`; reserve `Computed` for values returned by
  Nautobot or derived by the provider.
- Decide null, unknown, and empty-value behavior deliberately and cover it with tests. Missing
  nullable timestamps are e.g. represented as Terraform null values.
- Do not apply `UseStateForUnknown` to computed values that may change during an update. It is only
  appropriate when the previous value is guaranteed to remain valid.
- Avoid volatile metadata such as `last_updated` on resources when it adds state noise without
  affecting resource management. Such metadata is usually more useful on data sources.
- Keep matching single and list data sources aligned unless their documented semantics differ.

## Resource behavior

New resources should normally provide create, read, update, delete, and import support.

- Validate that create responses contain a non-empty object ID.
- During refresh, remove the resource from state when Nautobot returns not found.
- Treat deleting an already-absent object as success.
- Read the object after create and update so state reflects the API's canonical representation.
- Support clearing optional relationships as well as setting and replacing them.
- Do not silently ignore failures while resolving related objects or statuses.

## Data-source behavior

- A single-object data source should return a clear diagnostic for no match or malformed API data.
- List data sources must follow API pagination until all pages are read.
- Do not place list items with missing or empty IDs into state; return a diagnostic instead.
- Keep each list item schema aligned with the corresponding single-object data source.

## Testing conventions

Unit tests must not require a live Nautobot instance. Acceptance tests use the `TestAcc` prefix and
must generate unique object names so they can run concurrently.

Resource acceptance coverage should be proportional to the resource and normally include:

- Minimal and fully populated configurations.
- Updates followed by convergence to an empty plan.
- Import verification.
- Deletion and already-absent deletion behavior.
- Setting, replacing, and clearing relationships.
- External drift detection for managed attributes.

Every ordinary acceptance-test apply step automatically requires an empty follow-up plan unless
`ExpectNonEmptyPlan` is explicitly set. When testing external drift, keep the initial apply clean:
capture the object ID in the first step, mutate Nautobot in the next step's `PreConfig`, require a
non-empty plan, and then verify that a final apply restores the configured values.

List data-source tests should verify pagination, item identity, representative attributes, nullable
values, and relationships.

## Running acceptance tests

Start or reuse the local Nautobot stack and run the complete suite:

```sh
make testacc
```

Run tests directly against an already-running local instance:

```sh
make testacc-run
```

Run one test:

```sh
make testacc-run TEST=./internal/provider/... TESTARGS='-run TestAccTenantResource_drift'
```

The local containers deliberately remain running after tests finish or fail. Remove their containers
and volumes explicitly when finished:

```sh
make testacc-local-down
```

Override `NAUTOBOT_TEST_URL` and `NAUTOBOT_TEST_TOKEN` only when intentionally targeting another
disposable test instance. Select a CLI with `TF_TOOL=opentofu` or `TF_TOOL=terraform`.

## Commits and pull requests

- Write concise, imperative commit subjects that explain the change.
- Include tests with functional changes and explain any test that could not be run.
- Complete the pull request template, including compatibility and testing information.
- Call out breaking changes, schema changes, new dependencies, and Nautobot-version assumptions.
- Do not commit provider binaries, Terraform/OpenTofu state or plan files, local credentials, or
  acceptance-test artifacts.

By contributing, you agree that your contribution is licensed under the repository's
[Apache License 2.0](LICENSE).
