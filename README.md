# Sensu Go <HandlerName> Handler

The [Sensu Go][1] [HandlerName][2] handler is a [Sensu Event Handler][3] that performs something useful for Sensu.

## Configuration

Example Sensu Go handler definition:

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "handlername"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-handlername-handler -param1=value1",
        "timeout": 10,
        "filters": [
            "is_incident"
        ]
    }
}
```

## Usage Examples

This handler performs a task. 

**Help**

```
The Sensu Go handler for managing HandlerName data

Usage:
  sensu-handlername-handler [flags]

Flags:
  -p, --param1 string     The param1 usage information
```

**Invoke HandlerName to perform an operation**

Using environment variables
```bash
export HANDLER_PARAM1=value1
sensu-handlername-handler < event.json
```

Using command line arguments
```bash
sensu-handlername-handler -param1=value1 < event.json
```

## Contributing
See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://www.google.com
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://bitbucket.org/fguimond/sensu-handlername-handler/src
