package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/httpclient"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-aws-ec2-deregistration-handler/aws"
	"github.com/sensu/sensu-go/types"
)

const (
	keepAliveEventName = "keepalive"
)

var (
	awsConfig = aws.Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-aws-ec2-deregistration-handler",
			Short:    "removes sensu entities that do not have an allowed ec2 instance state",
			Timeout:  10,
			Keyspace: "sensu.io/plugins/ec2deregistration/config",
		},
	}

	awsInstanceIdLabel = ""

	sensuAPIURL string
	sensuAPIKey string
	sensuCACert string

	options = []*sensu.PluginConfigOption{
		{
			Path:      "aws-access-key-id",
			Env:       "AWS_ACCESS_KEY_ID",
			Argument:  "aws-access-key-id",
			Shorthand: "k",
			Default:   "",
			Usage:     "The AWS access key id to authenticate",
			Value:     &awsConfig.AwsAccessKeyId,
		},
		{
			Path:      "aws-secret-key",
			Env:       "AWS_SECRET_KEY",
			Argument:  "aws-secret-key",
			Shorthand: "s",
			Default:   "",
			Usage:     "The AWS secret key id to authenticate",
			Value:     &awsConfig.AwsSecretKey,
		},
		{
			Path:      "aws-instance-id",
			Env:       "AWS_INSTANCE_ID",
			Argument:  "aws-instance-id",
			Shorthand: "i",
			Default:   "",
			Usage:     "The AWS instance ID",
			Value:     &awsConfig.AwsInstanceId,
		},
		{
			Path:      "aws-instance-id-label",
			Env:       "AWS_INSTANCE_ID_LABEL",
			Argument:  "aws-instance-id-label",
			Shorthand: "l",
			Default:   "aws-instance-id",
			Usage:     "The entity label containing the AWS instance ID",
			Value:     &awsInstanceIdLabel,
		},
		{
			Path:      "aws-region",
			Env:       "AWS_REGION",
			Argument:  "aws-region",
			Shorthand: "r",
			Default:   "us-east-1",
			Usage:     "The AWS region",
			Value:     &awsConfig.AwsRegion,
		},
		{
			Path:      "aws-allowed-instance-states",
			Env:       "AWS_ALLOWED_INSTANCE_STATES",
			Argument:  "aws-allowed-instance-states",
			Shorthand: "S",
			Default:   "running",
			Usage:     "The EC2 instance states allowed",
			Value:     &awsConfig.AllowedInstanceStates,
		},
		{
			Path:      "timeout",
			Env:       "TIMEOUT",
			Argument:  "timeout",
			Shorthand: "t",
			Default:   uint64(10),
			Usage:     "The plugin timeout",
			Value:     &awsConfig.Timeout,
		},
		{
			Path:      "sensu-api-url",
			Env:       "SENSU_API_URL",
			Argument:  "sensu-api-url",
			Shorthand: "u",
			Default:   "http://localhost:8080",
			Usage:     "The Sensu API URL",
			Value:     &sensuAPIURL,
		},
		{
			Path:      "sensu-api-key",
			Env:       "SENSU_API_KEY",
			Argument:  "sensu-api-key",
			Shorthand: "a",
			Usage:     "The Sensu API key",
			Value:     &sensuAPIKey,
		},
		{
			Path:      "sensu-ca-cert",
			Env:       "SENSU_CA_CERT",
			Argument:  "sensu-ca-cert",
			Shorthand: "c",
			Usage:     "The Sensu Go CA Certificate",
			Value:     &sensuCACert,
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
	goHandler := sensu.NewGoHandler(&awsConfig.PluginConfig, options, checkArgs, executeHandler)
	goHandler.Execute()
}

// checkArgs is invoked by the go handler to perform validation of the values. If an error is returned
// the handler will not be executed.
func checkArgs(event *types.Event) error {
	retrieveAwsInstanceId(event)

	if len(awsConfig.AwsAccessKeyId) == 0 {
		return fmt.Errorf("aws-access-key-id must contain a value")
	}
	if len(awsConfig.AwsSecretKey) == 0 {
		return fmt.Errorf("aws-secret-key must contain a value")
	}
	if len(awsConfig.AwsInstanceId) == 0 {
		return fmt.Errorf("aws-instance-id must contain a value")
	}
	if len(awsConfig.AwsRegion) == 0 {
		return fmt.Errorf("aws-region must contain a value")
	}
	if len(awsConfig.AllowedInstanceStates) == 0 {
		return fmt.Errorf("allowed-instance-states must contain at least one value")
	}
	if len(sensuAPIURL) == 0 {
		return fmt.Errorf("sensu-api-url must contain a value")
	}
	_, err := url.Parse(sensuAPIURL)
	if err != nil {
		return fmt.Errorf("invalid value for sensu-api-url: %s", err)
	}
	if len(sensuAPIKey) == 0 {
		return fmt.Errorf("sensu-api-key must contain a value")
	}

	// parse the instance states
	awsConfig.AllowedInstanceStatesMap = make(map[string]bool)
	for _, instanceState := range strings.Split(awsConfig.AllowedInstanceStates, ",") {
		trimmedInstanceState := strings.TrimSpace(instanceState)
		if len(trimmedInstanceState) == 0 {
			// Ignore this one
		} else if validInstanceStates[trimmedInstanceState] {
			awsConfig.AllowedInstanceStatesMap[trimmedInstanceState] = true
		} else {
			return fmt.Errorf("invalid instance state: %s", trimmedInstanceState)
		}
	}

	return nil
}

// retrieveAwsInstanceId sets the AWS instance id using the entity label or entity name
// if the actual instance id is not set on the command line.
func retrieveAwsInstanceId(event *types.Event) {
	if len(awsConfig.AwsInstanceId) > 0 {
		return
	}

	if len(awsInstanceIdLabel) > 0 {
		if len(event.Entity.Labels[awsInstanceIdLabel]) > 0 {
			log.Printf("Using %s entity label as the AWS instance ID\n", awsInstanceIdLabel)
			awsConfig.AwsInstanceId = event.Entity.Labels[awsInstanceIdLabel]
		}
	}
	if len(awsConfig.AwsInstanceId) == 0 {
		log.Println("Using entity name as the AWS instance ID")
		awsConfig.AwsInstanceId = event.Entity.Name
	}
}

// executeHandler is executed by the go handler and executes the handler business logic.
func executeHandler(event *types.Event) error {
	if event.Check.Name != keepAliveEventName {
		return fmt.Errorf("received non-keepalive event, not checking ec2 instance state")
	}

	awsHandler, err := aws.NewHandler(&awsConfig)
	if err != nil {
		return fmt.Errorf("could not initialize handler: %s", err)
	}

	log.Println("Getting AWS instance state")
	instanceState, getErr := awsHandler.GetInstanceState()
	if getErr != nil {
		return fmt.Errorf("could not get instance state: %s", getErr)
	}

	log.Printf("Instance state: %s", instanceState)

	// Validate instance state
	if _, ok := awsConfig.AllowedInstanceStatesMap[instanceState]; ok {
		log.Printf("'%s' is a valid instance state, not deregistering '%s' entity from Sensu for '%s' AWS instance", instanceState,
			event.Entity.Name, awsConfig.AwsInstanceId)
		return nil
	}
	log.Printf("'%s' is not a valid instance state, deregistering '%s' entity from Sensu for '%s' AWS instance", instanceState,
		event.Entity.Name, awsConfig.AwsInstanceId)

	// First authenticate against the Sensu API
	config := httpclient.CoreClientConfig{
		URL:    sensuAPIURL,
		APIKey: sensuAPIKey,
	}
	if sensuCACert != "" {
		asn1Data, err := ioutil.ReadFile(sensuCACert)
		if err != nil {
			return fmt.Errorf("unable to load sensu-ca-cert: %s", err)
		}
		cert, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			return fmt.Errorf("invalid sensu-ca-cert: %s", err)
		}
		config.CACert = cert

	}
	client := httpclient.NewCoreClient(config)
	request, err := httpclient.NewResourceRequest("core/v2", "Entity", event.Entity.Namespace, event.Entity.Name)
	if err != nil {
		return err
	}

	// Delete the Sensu entity
	log.Printf("Deleting entity (%s/%s)", event.Entity.Namespace, event.Entity.Name)
	if _, err := client.DeleteResource(context.Background(), request); err != nil {
		if httperr, ok := err.(httpclient.HTTPError); ok {
			if httperr.StatusCode < 500 {
				log.Printf("entity already deleted (%s/%s)", event.Entity.Namespace, event.Entity.Name)
				return nil
			}
		}
		return err
	}

	return nil
}
