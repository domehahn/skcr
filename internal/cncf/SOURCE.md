# CNCF Landscape data

`landscape.yml` is an unmodified snapshot of the official CNCF Landscape data:

- Source: https://github.com/cncf/landscape/blob/master/landscape.yml
- Website: https://landscape.cncf.io/
- Retrieved: 2026-08-10
- Snapshot SHA-256: `c32564e7a64573273350e547d99adceb30ebe23727f5e0f280dcce2dbb747231`
- Rows: 2,413
- Unique item names represented as skills: 2,407

The CNCF repository makes `landscape.yml` alternatively available under the
Creative Commons Attribution 4.0 license. The source also contains data subject
to the Crunchbase Data Access Terms and restricted to Linux Foundation
landscape projects. SKCR retains the source attribution and does not embed
logos or claim CNCF endorsement.

Refresh the snapshot with:

```bash
go generate ./internal/cncf
```
