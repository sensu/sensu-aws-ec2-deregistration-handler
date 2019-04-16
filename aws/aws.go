package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sensu/sensu-enterprise-go-plugin/sensu"
	"log"
)

var (
	describeInstanceStatusIncludeAllInstances = true
)

type Config struct {
	sensu.HandlerConfig
	AwsAccessKeyId string
	AwsSecretKey   string
	AwsRegion      string
	AwsInstanceId  string
	//AwsAccounts           string
	AllowedInstanceStates string
	Timeout               uint64

	// Computed from the input
	AwsAccountsMap           map[string]bool
	AllowedInstanceStatesMap map[string]bool
}

type Handler struct {
	config     *Config
	awsSession *session.Session
	ec2Service *ec2.EC2
}

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
	creds := credentials.NewStaticCredentials(awsHandler.config.AwsAccessKeyId, awsHandler.config.AwsSecretKey, "")

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

func (awsHandler *Handler) GetInstanceState() (string, error) {
	instanceId := awsHandler.config.AwsInstanceId
	log.Printf("Retrieving instance state for %s\n", instanceId)

	request := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []*string{aws.String(instanceId)},
		IncludeAllInstances: &describeInstanceStatusIncludeAllInstances,
	}
	response, err := awsHandler.ec2Service.DescribeInstanceStatus(request)
	if err != nil {
		return "", fmt.Errorf("error getting instance state for %s: %s", instanceId, err)
	}
	log.Printf("DescribeInstanceStatus response: %s", response)

	instanceStatuses := response.InstanceStatuses
	if len(instanceStatuses) == 0 {
		return "", fmt.Errorf("could not get status for %s", instanceId)
	} else if len(instanceStatuses) > 1 {
		return "", fmt.Errorf("more than one instance found for %s", instanceId)
	}

	return *instanceStatuses[0].InstanceState.Name, nil
}

//func (awsHandler *Handler) AssumeRole() error {
//
//	credentials := sts.AssumeRoleInput{
//		RoleArn: awsHandler.
//	}
//	return nil
//}
