# victorialogs-fixture-seed

Seeds VictoriaLogs with a deterministic corpus of realistic, labeled log lines. One shot: generate, insert, exit.

The corpus imitates a small service landscape (web, api, db, auth across six hosts) with a realistic word distribution. Routine lines dominate, warnings are uncommon, and errors are rare, so a free-text LogsQL query like `refused` or `deadlock` returns a meaningful minority, the way grep over production logs would.

The same `-seed` and `-total` always produce the same lines, so any LogsQL count over the corpus reproduces across runs and machines. Only the timestamps differ between runs: they are spread over a window ending at run time, so the corpus lands in the recent past no matter when it is run.

## Scope

Content, not volume shape. The scenario is fixed and embedded, and the tool generates one batch and exits. Load generation, rate control, stream churn, and multi-tenant fan-out are out of scope.

## Tested with VictoriaLogs Version

This tool has been tested with VictoriaLogs v1.52.0 in cluster mode.

## Usage

Point it at a jsonline insert endpoint:

```
$ victorialogs-fixture-seed -addr http://127.0.0.1:9481/insert/jsonline
seeded 100000/1000000 lines to http://127.0.0.1:9481/insert/jsonline?_stream_fields=service%2Chost (window=24h0m0s seed=1)
seeded 200000/1000000 lines to http://127.0.0.1:9481/insert/jsonline?_stream_fields=service%2Chost (window=24h0m0s seed=1)
...
seeded 1000000/1000000 lines to http://127.0.0.1:9481/insert/jsonline?_stream_fields=service%2Chost (window=24h0m0s seed=1)
```

`service` and `host` are registered as log stream fields and `level` rides along as a regular field. An explicit `_stream_fields` query parameter in `-addr` takes precedence over that default.

Each generated line has this shape:

```
{"_time":"2026-07-31T04:59:29.662867888Z","_msg":"10.31.221.120 \"GET /static/app.js\" 200 78ms","level":"info","service":"web","host":"web-2"}
```

`-tenant` selects the tenant to ingest into, as `accountID` or `accountID:projectID` (sent as the `AccountID` and `ProjectID` request headers). Without it the corpus goes to the default tenant. One run seeds one tenant, and the same seed produces the same lines in whichever tenant it is pointed at, so give each tenant its own seed when the corpora should be distinguishable:

```
$ victorialogs-fixture-seed -addr http://127.0.0.1:9481/insert/jsonline -tenant 1 -seed 1
$ victorialogs-fixture-seed -addr http://127.0.0.1:9481/insert/jsonline -tenant 2 -seed 2
```

Run with `-help` for the full list of options.

## License

This project is licensed under the [MIT License](./LICENSE).
