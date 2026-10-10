#!/usr/bin/env python3
# Copyright Project Harbor Authors
# SPDX-License-Identifier: Apache-2.0
"""Compare issue #18856 before/after using identical real-client HTTP tests.

Usage: python3 tests/issue18856/measure.py --baseline <commit-before-fix> \
    --output /tmp/harbor-18856 --samples 100
Requires Go (per src/go.mod), Docker and the postgres:16-alpine image.
Starts an isolated PostgreSQL container on a random loopback port and stops it
in finally. No checkout, production service or existing database is modified.
The baseline changes only auth.go via a Go overlay; all tests/dependencies and
the rest of Harbor are identical. Token permissions, project lookups and the
downstream registry endpoint are fixtures, not a complete Harbor deployment.
"""

import argparse
from collections import Counter
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import platform
import statistics
import subprocess
import tempfile
import time


ROOT = Path(__file__).resolve().parents[2]
AUTH = "src/server/middleware/v2auth/auth.go"
SCENARIOS = {
    "pull", "head", "push", "upload", "delete_from_pull_token", "cross_project_mount",
    "sufficient_scope", "denied_scope",
}
BROKEN = SCENARIOS - {"sufficient_scope", "denied_scope"}


def command(args, **kwargs):
    return subprocess.check_output(args, cwd=ROOT, text=True, **kwargs).strip()


def summarize(rows, samples):
    counts = Counter(row["scenario"] for row in rows)
    if set(counts) != SCENARIOS or any(n != samples for n in counts.values()):
        raise RuntimeError(f"Incomplete measurements: {dict(counts)}")
    result = {}
    for scenario in sorted(SCENARIOS):
        group = [r for r in rows if r["scenario"] == scenario]
        durations = sorted(r["duration_ms"] for r in group)
        challenges = [e["challenge"].split(" ", 1)[0] for r in group
                      for e in r["events"] if e.get("challenge") and e["path"] != "/v2/"]
        result[scenario] = {
            "samples": len(group),
            "expected_status": group[0]["expected_status"],
            "expected_outcomes": sum(r["final_status"] == r["expected_status"] for r in group),
            "final_status_counts": dict(Counter(r["final_status"] for r in group)),
            "repository_challenge_schemes": dict(Counter(challenges)),
            "authorized_handler_calls": sum(r["authorized_handler_calls"] for r in group),
            "token_request_counts": dict(Counter(r["token_requests"] for r in group)),
            "session_p50_ms": round(statistics.median(durations), 3),
            "session_p95_ms": durations[max(0, (95 * len(durations) + 99) // 100 - 1)],
        }
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", required=True, help="Commit containing the old auth.go")
    parser.add_argument("--samples", type=int, default=100)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--postgres-image", default="postgres:16-alpine")
    args = parser.parse_args()
    if args.samples < 1:
        parser.error("--samples must be positive")
    out = args.output.resolve()
    out.mkdir(parents=True, exist_ok=True)
    if any((out / name).exists() for name in ("before.json", "after.json", "summary.json")):
        parser.error("Use a new output directory to preserve existing evidence")
    baseline_sha = command(["git", "rev-parse", args.baseline])
    container = f"harbor-18856-{os.getpid()}"
    started = False
    try:
        command(["docker", "run", "-d", "--rm", "--name", container,
                 "-p", "127.0.0.1::5432", "-e", "POSTGRES_PASSWORD=harbor-18856-test",
                 "-e", "POSTGRES_DB=registry", args.postgres_image])
        started = True
        db_port = command(["docker", "port", container, "5432/tcp"]).rsplit(":", 1)[1]
        for _ in range(150):
            ready = subprocess.run(["docker", "exec", container, "pg_isready", "-U", "postgres"],
                                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            if ready.returncode == 0:
                break
            time.sleep(0.2)
        else:
            raise RuntimeError("PostgreSQL did not become ready")
        env = dict(os.environ, POSTGRESQL_HOST="127.0.0.1", POSTGRESQL_PORT=db_port,
                   POSTGRESQL_USR="postgres", POSTGRESQL_PWD="harbor-18856-test",
                   POSTGRESQL_DATABASE="registry", HARBOR_ADMIN_PASSWD="Harbor12345",
                   POSTGRES_MIGRATION_SCRIPTS_PATH=str(ROOT / "make/migrations/postgresql"),
                   HARBOR_18856_SAMPLES=str(args.samples))
        results, exit_codes = {}, {}
        with tempfile.TemporaryDirectory(prefix="overlay-", dir=out) as tmp:
            tmp = Path(tmp)
            old_auth = tmp / "auth.go"
            old_auth.write_text(command(["git", "show", f"{baseline_sha}:{AUTH}"]) + "\n")
            overlay = tmp / "overlay.json"
            overlay.write_text(json.dumps({"Replace": {str(ROOT / AUTH): str(old_auth)}}))
            for phase in ("before", "after"):
                print(f"{phase}: {args.samples} sessions per scenario", flush=True)
                go_args = ["go", "test", "./server/middleware/v2auth", "-count=1", "-v"]
                if phase == "before":
                    go_args += [f"-overlay={overlay}"]
                with (out / f"{phase}.log").open("w") as log:
                    run = subprocess.run(go_args, cwd=ROOT / "src", stdout=log,
                                         stderr=subprocess.STDOUT,
                                         env=dict(env, HARBOR_18856_RESULTS=str(out / f"{phase}.json")))
                exit_codes[phase] = run.returncode
                rows = json.loads((out / f"{phase}.json").read_text())
                results[phase] = summarize(rows, args.samples)
        reproduced = exit_codes["before"] != 0 and all(
            results["before"][s]["expected_outcomes"] == 0
            and results["before"][s]["final_status_counts"] == {401: args.samples}
            and results["before"][s]["repository_challenge_schemes"] == {"Basic": 2 * args.samples}
            for s in BROKEN)
        fixed = exit_codes["after"] == 0 and all(
            row["expected_outcomes"] == args.samples for row in results["after"].values())
        denied_safe = all(results[p]["denied_scope"]["authorized_handler_calls"] == 0
                          and results[p]["denied_scope"]["final_status_counts"] == {401: args.samples}
                          for p in ("before", "after"))
        fingerprints = {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest()
                        for p in (ROOT / AUTH, ROOT / "src/server/middleware/v2auth/challenge_test.go",
                                  ROOT / "src/server/middleware/v2auth/token_refresh_test.go")}
        summary = {
            "issue": "https://github.com/goharbor/harbor/issues/18856",
            "timestamp_utc": datetime.now(timezone.utc).isoformat(),
            "baseline_commit": baseline_sha, "candidate_sha256": fingerprints,
            "go_version": command(["go", "version"]), "platform": platform.platform(),
            "postgres_image": args.postgres_image,
            "scope": "Real go-containerregistry transport over TLS; actual Harbor artifact, JWT and auth middleware; fixture token ACL, project lookups and registry handler",
            "latency_note": "Local protocol-session diagnostics, including ping and token exchange; not production latency or a throughput benchmark",
            "test_exit_codes": exit_codes, "baseline_reproduced": reproduced,
            "candidate_passed": fixed, "denied_access_preserved": denied_safe,
            **results,
        }
        (out / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
        print("scenario                         before       after")
        for scenario in sorted(SCENARIOS):
            before = results["before"][scenario]["expected_outcomes"]
            after = results["after"][scenario]["expected_outcomes"]
            print(f"{scenario:32} {before:4}/{args.samples:<5} {after:4}/{args.samples}")
        print(f"Evidence: {out / 'summary.json'}")
        if not (reproduced and fixed and denied_safe):
            raise SystemExit("Verification failed; inspect before.log, after.log and summary.json")
    finally:
        if started:
            subprocess.run(["docker", "stop", container], check=True, stdout=subprocess.DEVNULL)


if __name__ == "__main__":
    main()
