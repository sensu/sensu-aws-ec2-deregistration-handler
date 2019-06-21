# Sensu Go <HandlerName> Handler

The [Sensu Go][1] AWS EC2 Deregistration handler is a [Sensu Event Handler][2] that checks an AWS EC2 instance and removes it from Sensu if it is not in one of the specified state.

## Configuration

Example Sensu Go handler definition:

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "awsEc2Deregistration"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-aws-ec2-deregistration-handler -aws-access-key-id=aaaa -aws-secret-key=key -aws-region=us-east-2 -aws-instance-id-label=aws-instance-id -aws-allowed-instance-states=running,stopped,stopping -sensu-api-url=http://localhost:8080 -sensu-api-username=admin -sensu-api-password=password",
        "timeout": 10,
        "filters": [
            "is_incident"
        ]
    }
}
```

## Usage Examples

This handler checks an AWS EC2 instance and removes it from Sensu if it is not in one of the specified state.

The AWS EC2 instance ID can be read either as a handler option, or using an entity label specified in the `The AWS EC2 instance ID can be read either as a handler option, or using an entity label specified in the `aws-instance-id-label` option.

**Help**

```
removes sensu clients that do not have an allowed ec2 instance state

Usage:
  sensu-aws-ec2-deregistration-handler [flags]

Flags:
  -k, --aws-access-key-id string             The AWS access key id to authenticate
  -S, --aws-allowed-instance-states string   The EC2 instance states allowed (default "running")
  -i, --aws-instance-id string               The AWS instance ID
  -l, --aws-instance-id-label string         The entity label containing the AWS instance ID
  -r, --aws-region string                    The AWS region (default "us-east-1")
  -s, --aws-secret-key string                The AWS secret key id to authenticate
  -h, --help                                 help for sensu-aws-ec2-deregistration-handler
  -p, --sensu-api-password string            The Sensu API password
  -U, --sensu-api-url string                 The Sensu API URL (default "http://localhost:8080")
  -u, --sensu-api-username string            The Sensu API username
  -t, --timeout uint                         The plugin timeout (default 10)```
```

Using environment variables
```bash
export AWS_ACCESS_KEY_ID=acesskey
export AWS_SECRET_KEY=secretkey
export AWS_REGION=us-east-2
export AWS_INSTANCE_ID_LABEL=aws-instance-id
export AWS_ALLOWED_INSTANCE_STATES=running,stopped,stopping
export SENSU_API_URL=http://localhost:8080export 
export SENSU_API_USERNAME=admin
export SENSU_API_PASSWORD=password
sensu-aws-ec2-deregistration-handler < event.json
```

Using command line arguments
```bash
sensu-aws-ec2-deregistration-handler -aws-access-key-id=aaaa -aws-secret-key=key -aws-region=us-east-2 -aws-instance-id-label=aws-instance-id -aws-allowed-instance-states=running,stopped,stopping -sensu-api-url=http://localhost:8080 -sensu-api-username=admin -sensu-api-password=password < event.json
```

## Contributing
See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[3]: https://github.com/sensu-skunkworks/sensu-aws-ec2-deregistration-handler/src
