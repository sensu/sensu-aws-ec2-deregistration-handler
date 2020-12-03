[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-ec2-handler)

# Sensu Go EC2 Handler

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
  - [Environment variables](#environment-variables)
  - [Annotations](#annotations)
  - [AWS Credentials](#aws-credentials)
  - [Proxy support](#proxy-support)
- [Installation from source](#installation-from-source)
- [Contributing](#contributing)

## Overview

The [Sensu Go][1] EC2 handler is a [Sensu Event Handler][2] that checks an AWS
EC2 instance and removes it from Sensu if it is not in one of the specified
state.

## Usage Examples

This handler checks an AWS EC2 instance and removes it from Sensu if it is not
in one of the specified state.

The AWS EC2 instance ID can be read either as a handler option, or using an
entity label specified in the `aws-instance-id-label` option.

### Help

```
removes sensu clients that do not have an allowed ec2 instance state

Usage:
  sensu-ec2-handler [flags]

Flags:
  -k, --aws-access-key-id string             The AWS access key id to authenticate
  -s, --aws-secret-key string                The AWS secret key id to authenticate
  -S, --aws-allowed-instance-states string   The EC2 instance states allowed (default "running")
  -i, --aws-instance-id string               The AWS instance ID
  -l, --aws-instance-id-label string         The entity label containing the AWS instance ID
  -r, --aws-region string                    The AWS region (default "us-east-1")
  -R, --aws-assume-role-arn string           The AWS IAM Role to assume, if necessary
  -U, --sensu-api-url string                 The Sensu API URL (default "http://localhost:8080")
  -a, --sensu-api-key string                 The Sensu API key
  -c, --sensu-ca-cert string                 The Sensu Go CA Certificate
  -t, --timeout uint                         The plugin timeout (default 10)```
  -h, --help                                 help for sensu-ec2-handler
```

## Configuration

### Asset registration

[Sensu Assets][4] are the best way to make use of this plugin. If you're not
using an asset, please consider doing so! If you're using sensuctl 5.13 with
Sensu Backend 5.13 or later, you can use the following command to add the asset:

```
sensuctl asset add sensu/sensu-ec2-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the
[Bonsai Asset Index][3].

### Handler definition

Example Sensu Go handler definition:

```yaml
type: Handler
api_version: core/v2
metadata:
  namespace: default
  name: sensu-ec2-handler
spec:
  type: pipe
  runtime_assets:
    - sensu/sensu-ec2-handler
  filters:
    - is_incident
    - not_silenced
  command: >-
    sensu-ec2-handler
    --aws-region us-east-2
    --aws-instance-id-label aws-instance-id
    --aws-allowed-instance-states running,stopped,stopping
    --sensu-api-url http://localhost:8080
  secrets:
    - name: AWS_ACCESS_KEY_ID
      secret: aws_access_key_id
    - name: AWS_SECRET_KEY
      secret: aws_secret_key
    - name: SENSU_API_KEY
      secret: sensu_api_key
```

### EC2 instance states

The `--aws-allowed-instance-states` argument allows you to specify the valid
states for an EC2 instance to be in.  If the instance in the Sensu event is not
in one of these states, it will be deregistered from Sensu.

The available instance states are:
* pending
* running
* stopping
* stopped
* shutting-down
* terminated

### Environment variables

Most arguments for this handler are available to be set via environment
variables.  However, any arguments specified directly on the command line
override the corresponding environment variable.


|Argument                     |Environment Variable       |
|-----------------------------|---------------------------|
|--aws-access-key-id          |AWS_ACCESS_KEY_ID          |
|--aws-secret-key             |AWS_SECRET_KEY             |
|--aws-region                 |AWS_REGION                 |
|--aws-instance-id            |AWS_INSTANCE_ID            |
|--aws-instance-id-label      |AWS_INSTANCE_ID_LABEL      |
|--aws-allowed-instance-states|AWS_ALLOWED_INSTANCE_STATES|
|--aws-assume-role-arn        |AWS_ASSUME_ROLE_ARN        |
|--sensu-api-url              |SENSU_API_URL              |
|--sensu-api-key              |SENSU_API_KEY              |
|--sensu-ca-cert              |SENSU_CA_CERT              |
|--timeout                    |TIMEOUT                    |

**Security Note:** Care should be taken to not expose the AWS access and secret
keys or the Sensu API key information for this handler by specifying them on
the command line or by directly setting the environment variables in the handler
definition.  It is suggested to make use of [secrets management][5] to surface
them as environment variables.  The handler definition above references them as
secrets.  Below is an example secrets definition that make use of the built-in
[env secrets provider][6].

```yml
---
type: Secret
api_version: secrets/v1
metadata:
  name: aws_secret_key
spec:
  provider: env
  id: AWS_SECRET_KEY
---
type: Secret
api_version: secrets/v1
metadata:
  name: aws_access_key_id
spec:
  provider: env
  id: AWS_ACCESS_KEY_ID
---
type: Secret
api_version: secrets/v1
metadata:
  name: sensu_api_key
spec:
  provider: env
  id: SENSU_API_KEY
```

### Annotations

All arguments for this handler are tunable on a per entity or check basis based
on annotations.  The annotations keyspace for this handler is
`sensu.io/plugins/sensu-ec2-handler/config`.

###  AWS Credentials

**NOTE:** Providing AWS credentials via the command line arguments `--aws-access-key-id` and
`--aws-secret-key` is deprecated and will be removed in a future release.  Please use one
of the methods below.

This plugin makes use of the AWS SDK for Go.  The SDK uses the [default credential provider chain][7]
to find AWS credentials.  The SDK uses the first provider in the chain that returns credentials
without an error. The default provider chain looks for credentials in the following order:

1. Environment variables (AWS_SECRET_ACCESS_KEY, AWS_ACCESS_KEY_ID, and AWS_REGION).

2. Shared credentials file (typically ~/.aws/credentials).

3. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.

4. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.

The SDK detects and uses the built-in providers automatically, without requiring manual configurations.
For example, if you use IAM roles for Amazon EC2 instances, your applications automatically use the
instance’s credentials. You don’t need to manually configure credentials in your application.

Source: [Configuring the AWS SDK for Go][8]

This plugin also supports assuming a new role upon authentication using the `--aws-assume-role-arn`
option.

If you go the route of using environment variables, it is highly suggested you use them via the
[Env secrets provider][6].

### Proxy Support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is
assumed.

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an
Asset. If you would like to compile and install the plugin from source or
contribute to it, download the latest version or create an executable binary
from this source.

From the local path of the sensu-ec2-handler repository:

```
go build
```

## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://docs.sensu.io/sensu-go/latest/reference/handlers/#how-do-sensu-handlers-work
[3]: https://bonsai.sensu.io/assets/sensu/sensu-ec2-handler
[4]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[5]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/
[6]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/#use-env-for-secrets-management
[7]: https://docs.aws.amazon.com/sdk-for-go/api/aws/defaults/#CredChain
[8]: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html

