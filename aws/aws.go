package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	var err error
	creds := credentials.NewStaticCredentials(awsHandler.config.AwsAccessKeyID, awsHandler.config.AwsSecretKey, "")

	awsHandler.awsSession, err = session.NewSession(&aws.Config{
		Region:      aws.String(awsHandler.config.AwsRegion),
		Credentials: creds,
	})
	if err != nil {
		return err
	}
	log.Println("Session created!")

	awsHandler.ec2Service = ec2.New(awsHandler.awsSession)

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
