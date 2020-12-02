package aws

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
)

var (
	describeInstanceStatusIncludeAllInstances = true
)

// Config is the aws config
type Config struct {
	sensu.PluginConfig
	AwsAccessKeyID        string
	AwsSecretKey          string
	AwsRegion             string
	AwsInstanceID         string
	AllowedInstanceStates string
	Timeout               uint64
	RoleArn               string

	// Computed from the input
	AwsAccountsMap           map[string]bool
	AllowedInstanceStatesMap map[string]bool
}

// Handler is the aws handler
type Handler struct {
	config     *Config
	awsSession *session.Session
	ec2Service *ec2.EC2
}

// NewHandler creates a new handler
func NewHandler(config *Config) (*Handler, error) {
	handler := Handler{
		config: config,
	}

	err := handler.initAws()
	if err != nil {
		return nil, fmt.Errorf("error initializing aws handler: %s", err)
	}

	return &handler, nil
}

func (awsHandler *Handler) initAws() error {
	log.Println("Creating AWS session...")

	// to prevent a breaking change of supplying the AWS Creds/Region via
	// arguments, but not in the env, push them back to the env to be picked
	// up by the session
	if len(awsHandler.config.AwsAccessKeyID) > 0 {
		os.Setenv("AWS_ACCESS_KEY_ID", awsHandler.config.AwsAccessKeyID)
	}
	if len(awsHandler.config.AwsSecretKey) > 0 {
		os.Setenv("AWS_SECRET_KEY", awsHandler.config.AwsSecretKey)
	}
	if len(awsHandler.config.AwsRegion) > 0 {
		os.Setenv("AWS_REGION", awsHandler.config.AwsRegion)
	}

	awsHandler.awsSession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	log.Println("Session created!")

	if arn.IsARN(awsHandler.config.RoleArn) {
		log.Println("Using Role ARN")
		creds := stscreds.NewCredentials(awsHandler.awsSession, awsHandler.config.RoleArn)
		awsHandler.ec2Service = ec2.New(awsHandler.awsSession, &aws.Config{Credentials: creds})
	} else {
		awsHandler.ec2Service = ec2.New(awsHandler.awsSession)
	}


	return nil
}

// GetInstanceState gets the instance state
func (awsHandler *Handler) GetInstanceState() (string, error) {
	instanceID := awsHandler.config.AwsInstanceID
	log.Printf("Retrieving AWS instance state for %s\n", instanceID)

	request := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []*string{aws.String(instanceID)},
		IncludeAllInstances: &describeInstanceStatusIncludeAllInstances,
	}
	response, err := awsHandler.ec2Service.DescribeInstanceStatus(request)
	if err != nil {
		return "", fmt.Errorf("error getting instance state for %s: %s", instanceID, err)
	}

	instanceStatuses := response.InstanceStatuses
	if len(instanceStatuses) == 0 {
		return "", fmt.Errorf("could not get status for %s", instanceID)
	} else if len(instanceStatuses) > 1 {
		return "", fmt.Errorf("more than one instance found for %s", instanceID)
	}

	return *instanceStatuses[0].InstanceState.Name, nil
}
