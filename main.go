package main

import (
	"encoding/json"
	"fmt"
	"github.com/sensu/sensu-enterprise-go-plugin/sensu"
	"github.com/sensu/sensu-go/types"
	"log"
	"strings"
)

type Config struct {
	sensu.HandlerConfig
	AwsAccessKeyId        string
	AwsSecretKey          string
	AwsRegion             string
	AwsAccounts           string
	AllowedInstanceStates string
	Timeout               uint64

	// Computed from the input
	AwsAccountsMap           map[string]bool
	AllowedInstanceStatesMap map[string]bool
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
			Path:      "aws-access-key-id",
			Env:       "AWS_ACCESS_KEY_ID",
			Argument:  "aws-access-key-id",
			Shorthand: "k",
			Default:   "",
			Usage:     "The AWS access key id to authenticate",
			Value:     &config.AwsAccessKeyId,
		},
		{
			Path:      "aws-secret-key",
			Env:       "AWS_SECRET_KEY",
			Argument:  "aws-secret-key",
			Shorthand: "s",
			Default:   "",
			Usage:     "The AWS secret key id to authenticate",
			Value:     &config.AwsSecretKey,
		},
		{
			Path:      "aws-region",
			Env:       "AWS_REGION",
			Argument:  "aws-region",
			Shorthand: "r",
			Default:   "us-east-1",
			Usage:     "The AWS region",
			Value:     &config.AwsRegion,
		},
		{
			Path:      "aws-allowed-instance-states",
			Env:       "AWS_ALLOWED_INSTANCE_STATES",
			Argument:  "aws-allowed-instance-states",
			Shorthand: "S",
			Default:   "running",
			Usage:     "The EC2 instance states allowed",
			Value:     &config.AllowedInstanceStates,
		},
		{
			Path:      "aws-accounts",
			Env:       "AWS_ACCOUNTS",
			Argument:  "aws-accounts",
			Shorthand: "a",
			Default:   "",
			Usage:     "The AWS accounts",
			Value:     &config.AwsAccounts,
		},
		{
			Path:      "timeout",
			Env:       "TIMEOUT",
			Argument:  "timeout",
			Shorthand: "t",
			Default:   uint64(10),
			Usage:     "The plugin timeout",
			Value:     &config.Timeout,
		},
	}

	validInstanceStates = map[string]bool{
		"pending":       true,
		"running":       true,
		"stopping":      true,
		"stopped":       true,
		"shutting-down": true,
		"terminated":    true,
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
	if len(config.AwsRegion) == 0 {
		return fmt.Errorf("aws-region must contain a value")
	}
	if len(config.AwsAccessKeyId) == 0 {
		return fmt.Errorf("aws-access-key-id must contain a value")
	}
	if len(config.AwsSecretKey) == 0 {
		return fmt.Errorf("aws-secret-key must contain a value")
	}
	if len(config.AwsAccounts) == 0 {
		return fmt.Errorf("aws-accounts must contain at least one value")
	}
	if len(config.AllowedInstanceStates) == 0 {
		return fmt.Errorf("allowed-instance-states must contain at least one value")
	}

	// parse the aws accounts
	config.AwsAccountsMap = make(map[string]bool)
	for _, account := range strings.Split(config.AwsAccounts, ",") {
		trimmedAccount := strings.TrimSpace(account)
		if len(trimmedAccount) > 0 {
			config.AwsAccountsMap[trimmedAccount] = true
		}
	}

	// parse the instance states
	config.AllowedInstanceStatesMap = make(map[string]bool)
	for _, instanceState := range strings.Split(config.AllowedInstanceStates, ",") {
		trimmedInstanceState := strings.TrimSpace(instanceState)
		if len(trimmedInstanceState) == 0 {
			// Ignore this one
		} else if validInstanceStates[trimmedInstanceState] {
			config.AllowedInstanceStatesMap[trimmedInstanceState] = true
		} else {
			return fmt.Errorf("invalid instance state: %s", trimmedInstanceState)
		}
	}

	return nil
}

// executeHandler is executed by the go handler and executes the handler business logic.
func executeHandler(event *types.Event) error {
	configJsonBytes, _ := json.MarshalIndent(config, "", "    ")
	log.Printf("Config:\n%s\n", string(configJsonBytes))
	jsonBytes, _ := json.MarshalIndent(event, "", "    ")
	log.Printf("Event:\n%s\n", string(jsonBytes))

	return nil
}
