This directory contains the clean development baseline. `000001_initial_schema` creates the final schema directly; it does not replay historic role, notification, payment, or appointment patches.

This intentionally replaces the prior migration history. Existing development databases must be recreated before running it:

```bash
docker compose down -v
docker compose up --build
```

The baseline does not include demo data. Add future changes as a new numbered migration; do not edit `000001_initial_schema` after it has been shared.
