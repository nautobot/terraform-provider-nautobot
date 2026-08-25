## Summary

<!-- Describe what this PR changes and why. Focus on user-visible behavior. -->

## Related issue

<!-- Use "Closes #123" when this PR should close an issue. Write "None" if not applicable. -->

## Type of change

- [ ] Bug fix
- [ ] New or changed resource/data source
- [ ] Internal refactor
- [ ] Tests
- [ ] Documentation
- [ ] Dependency or build change
- [ ] Breaking change

## Implementation notes

<!--
Describe important design decisions, public schema changes, compatibility considerations,
null/empty semantics, API-version assumptions, or follow-up work.
-->

## Testing

<!-- List the commands and relevant versions used. Explain anything that was not run. -->

- [ ] `make fmt-check`
- [ ] `make test`
- [ ] Relevant acceptance tests with OpenTofu
- [ ] Relevant acceptance tests with Terraform
- [ ] `make docs` when provider schemas or descriptions changed

Commands/results:

```text

```

## Checklist

- [ ] I read and followed [CONTRIBUTORS.md](../CONTRIBUTORS.md) and the [Code of Conduct](CODE_OF_CONDUCT.md).
- [ ] The change is focused and does not include unrelated formatting or generated artifacts.
- [ ] Public attribute names and existing state behavior remain backward compatible, or the breaking change is clearly documented.
- [ ] New behavior has appropriate unit and/or acceptance coverage.
- [ ] Affected resource apply steps converge to an empty follow-up plan.
- [ ] Affected resource reads and deletes handle objects that are already absent.
- [ ] Documentation and examples are updated where users will observe a change.
- [ ] No credentials, tokens, state files, plans, or provider binaries are included.
