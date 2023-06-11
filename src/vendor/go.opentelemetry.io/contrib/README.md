# OpenTelemetry-Go Contrib

[![build_and_test](https://github.com/open-telemetry/opentelemetry-go-contrib/workflows/build_and_test/badge.svg)](https://github.com/open-telemetry/opentelemetry-go-contrib/actions?query=workflow%3Abuild_and_test+branch%3Amain)
[![codecov.io](https://codecov.io/gh/open-telemetry/opentelemetry-go-contrib/coverage.svg?branch=main)](https://app.codecov.io/gh/open-telemetry/opentelemetry-go-contrib?branch=main)
[![Docs](https://godoc.org/go.opentelemetry.io/contrib?status.svg)](https://pkg.go.dev/go.opentelemetry.io/contrib)
[![Go Report Card](https://goreportcard.com/badge/go.opentelemetry.io/contrib)](https://goreportcard.com/report/go.opentelemetry.io/contrib)
[![Slack](https://img.shields.io/badge/slack-@cncf/otel--go-brightgreen.svg?logo=slack)](https://cloud-native.slack.com/archives/C01NPAXACKT)

Collection of 3rd-party packages for [OpenTelemetry-Go](https://github.com/open-telemetry/opentelemetry-go).

## Contents

- [Instrumentation](./instrumentation/): Packages providing OpenTelemetry instrumentation for 3rd-party libraries.
- [Propagators](./propagators/): Packages providing OpenTelemetry context propagators for 3rd-party propagation formats.
- [Detectors](./detectors/): Packages providing OpenTelemetry resource detectors for 3rd-party cloud computing environments.

## Project Status

This project is currently in a pre-GA phase. Our progress towards a GA release
candidate is tracked in [this project
board](https://github.com/orgs/open-telemetry/projects/5).

### Compatibility

OpenTelemetry-Go Contrib attempts to track the current supported versions of the
[Go language](https://golang.org/doc/devel/release#policy). The release
schedule after a new minor version of go is as follows:

- The first release or one month, which ever is sooner, will add build steps for the new go version.
- The first release after three months will remove support for the oldest go version.

This project is tested on the following systems.

| OS      | Go Version | Architecture |
| ------- | ---------- | ------------ |
| Ubuntu  | 1.18       | amd64        |
| Ubuntu  | 1.17       | amd64        |
| Ubuntu  | 1.16       | amd64        |
| Ubuntu  | 1.18       | 386          |
| Ubuntu  | 1.17       | 386          |
| Ubuntu  | 1.16       | 386          |
| MacOS   | 1.18       | amd64        |
| MacOS   | 1.17       | amd64        |
| MacOS   | 1.16       | amd64        |
| Windows | 1.18       | amd64        |
| Windows | 1.17       | amd64        |
| Windows | 1.16       | amd64        |
| Windows | 1.18       | 386          |
| Windows | 1.17       | 386          |
| Windows | 1.16       | 386          |

While this project should work for other systems, no compatibility guarantees
are made for those systems currently.

Go 1.18 was added in March of 2022.
Go 1.16 will be removed around June 2022.

## Contributing

For information on how to contribute, consult [the contributing guidelines](./CONTRIBUTING.md)
