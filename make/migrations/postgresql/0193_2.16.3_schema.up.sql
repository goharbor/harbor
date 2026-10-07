/*
Dockerfile generation fallback: when an artifact carries no BuildKit provenance
attestation, the optimizer adapter reconstructs a best-effort Dockerfile from
the image's own config history instead of failing outright. `generated`
distinguishes these approximate reconstructions from Dockerfiles extracted
verbatim from provenance. Existing rows were all produced by the extraction
path, so they backfill as FALSE.
*/
ALTER TABLE dockerfile_optimization
    ADD COLUMN generated boolean NOT NULL DEFAULT FALSE;
