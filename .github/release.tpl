// publish release writes the following to release notes.
{"buildNo":"${buildNo}","preTag":"${preTag}"}
// buildNo is required,buildNo is used to specify the installer package.
// buildNo is found at https://github.com/goharbor/harbor/actions/workflows/build-package.yml
// preTag is not required, preTag is used to specify the previous tag, and generates release notes between previous tag and current tag. If preTag is not specified, previous tag will be automatically calculated.

// Example 1
// current tag:v2.5.0-rc1
{"buildNo":"build.1250","preTag":"v2.4.0"}
// Specify the installer package built by Build Package Workflow #1272 as release asset.
// Specify the previous tag as v2.4.0

// Example 2
// current tag:v2.5.0
{"buildNo":"rc1"}
// Specify the installer package of v2.5.0-rc1 as the release asset of v2.5.0.
// Unspecified preTag automatically calculates preTag for v2.4.0

// If the wrong buildNo is specified and the Workflow fails to run, please modify the buildNo and re run