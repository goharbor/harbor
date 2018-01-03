<!--[metadata]>
+++
title = "Common Server and Signer Configurations"
description = "Configuring the notary client, server and signer."
keywords = ["docker, notary, notary-client, notary-server, notary server, notary-signer, notary signer"]
[menu.main]
parent="mn_notary_config"
weight=5
+++
<![end-metadata]-->

# Configure sections common to Notary server and signer

The logging and bug reporting configuration options for both Notary server and
Notary signer have the same keys and format. The following sections provide
further detail.

For full specific configuration information, see the configuration files for the
Notary [server](server-config.md) or [signer](signer-config.md).

## logging section (optional)

The logging section sets the log level of the server.  If it is not provided,
the signer/server defaults to an ERROR logging level.  However if an explicit
value was provided, it must be a valid value.

Example:

```json
"logging": {
  "level": "debug"
}
```

Note that this entire section is optional.  However, if you would like to
specify a different log level, then you need the required parameters
below to configure it.

<table>
	<tr>
		<th>Parameter</th>
		<th>Required</th>
		<th>Description</th>
	</tr>
	<tr>
		<td valign="top"><code>level</code></td>
		<td valign="top">yes</td>
		<td valign="top">One of <code>"debug"</code>, <code>"info"</code>,
			<code>"warning"</code>, <code>"error"</code>, <code>"fatal"</code>,
			or <code>"panic"</code></td>
	</tr>
</table>

## reporting section (optional)

The reporting section contains any configuration for useful for running the
service, such as reporting errors. Currently, Notary only supports reporting errors
to <a href="https://bugsnag.com" target="_blank">Bugsnag</a>.

See <a href="https://github.com/bugsnag/bugsnag-go/" target="_blank">bugsnag-go</a> for more information
about these configuration parameters.

```json
"reporting": {
  "bugsnag": {
    "api_key": "c9d60ae4c7e70c4b6c4ebd3e8056d2b8",
    "release_stage": "production"
  }
}
```

Note that this entire section is optional.  If you want to report errors to
Bugsnag, then you need to include a `bugsnag` subsection, along with the
required parameters below, to configure it.

**Bugsnag reporting:**

<table>
	<tr>
		<th>Parameter</th>
		<th>Required</th>
		<th>Description</th>
	</tr>
	<tr>
		<td valign="top"><code>api_key</code></td>
		<td valign="top">yes</td>
		<td>The BugSnag API key to use to report errors.</td>
	</tr>
	<tr>
		<td valign="top"><code>release_stage</code></td>
		<td valign="top">yes</td>
		<td>The current release stage, such as <code>"production"</code>.  You can
			use this value to filter errors in the Bugsnag dashboard.</td>
	</tr>
</table>

## Related information

* [Notary Server Configuration File](server-config.md)
* [Notary Signer Configuration File](signer-config.md)
