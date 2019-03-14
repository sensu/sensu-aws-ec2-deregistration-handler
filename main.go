package main

import (
	"encoding/json"
	"fmt"
	"github.com/sensu/sensu-enterprise-go-plugin/sensu"
	"github.com/sensu/sensu-go/types"
	"log"
)

type Config struct {
	sensu.HandlerConfig
	param1 string
}

var (
	config = Config{
		HandlerConfig: sensu.HandlerConfig{
			Name:     "sensu-handler",
			Short:    "Sensu Go handler",
			Timeout:  10,
			Keyspace: "sensu.io/plugins/ec2deregistration/config",
		},
	}

	options = []*sensu.HandlerConfigOption{
		{
			Path:      "param1",
			Env:       "HANDLER_PARAM1",
			Argument:  "param1",
			Shorthand: "p",
			Default:   "defaultvalue",
			Usage:     "The param1 usage information",
			Value:     &config.param1,
		},
	}
)

func main() {
	goHandler := sensu.NewGoHandler(&config.HandlerConfig, options, checkArgs, executeHandler)
	err := goHandler.Execute()
	if err != nil {
		fmt.Printf("Error executing plugin: %s", err)
	}
}

// checkArgs is invoked by the go handler to perform validation of the values. If an error is returned
// the handler will not be executed.
func checkArgs(_ *types.Event) error {
	if len(config.param1) == 0 {
		return fmt.Errorf("param1 must contain a value")
	}

	return nil
}

// executeHandler is executed by the go handler and executes the handler business logic.
func executeHandler(event *types.Event) error {
	log.Printf("param1: %s\n", config.param1)
	jsonBytes, _ := json.Marshal(event)
	log.Printf("Event:%s\n", string(jsonBytes))

	return nil
}
