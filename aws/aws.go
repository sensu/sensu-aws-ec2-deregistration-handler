package aws

import (
	"fmt"
	"log"
        "time"

        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/credentials"
        "github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
        "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
        "github.com/aws/aws-sdk-go/aws/ec2metadata"
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/ec2"
        "github.com/sensu-community/sensu-plugin-sdk/sensu"
)

var (
	describeInstanceStatusIncludeAllInstances = true
)

type Config struct {
	sensu.PluginConfig
	AwsAccessKeyId string
	AwsSecretKey   string
	AwsRegion      string
	AwsInstanceId  string
        SensuRoleArn   string

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
	awsHandler.awsSession, err = session.NewSession(&aws.Config{
		Region:      aws.String(awsHandler.config.AwsRegion),
	})
	if err != nil {
		return err
	}
	log.Println("Session created!")

        creds := credentials.NewChainCredentials([]credentials.Provider{
            &credentials.EnvProvider{},
            &credentials.StaticProvider{
                Value: credentials.Value{
                    AccessKeyID:     awsHandler.config.AwsAccessKeyId,
                    SecretAccessKey: awsHandler.config.AwsSecretKey,
                },
            },
            &ec2rolecreds.EC2RoleProvider{
                Client:       ec2metadata.New(awsHandler.awsSession),
                ExpiryWindow: 5 * time.Minute,
            },
        })
        awsHandler.awsSession.Config.Credentials = creds

        if(len(awsHandler.config.SensuRoleArn) > 0){
            creds := stscreds.NewCredentials(awsHandler.awsSession,
                                             awsHandler.config.SensuRoleArn)
            awsHandler.awsSession.Config.Credentials = creds
        }

	awsHandler.ec2Service = ec2.New(awsHandler.awsSession)

	return nil
}

func (awsHandler *Handler) GetInstanceState() (string, error) {
	instanceId := awsHandler.config.AwsInstanceId
	log.Printf("Retrieving AWS instance state for %s\n", instanceId)

	request := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []*string{aws.String(instanceId)},
		IncludeAllInstances: &describeInstanceStatusIncludeAllInstances,
	}
	response, err := awsHandler.ec2Service.DescribeInstanceStatus(request)
	if err != nil {
		return "", fmt.Errorf("error getting instance state for %s: %s", instanceId, err)
	}

	instanceStatuses := response.InstanceStatuses
	if len(instanceStatuses) == 0 {
		return "", fmt.Errorf("could not get status for %s", instanceId)
	} else if len(instanceStatuses) > 1 {
		return "", fmt.Errorf("more than one instance found for %s", instanceId)
	}

	return *instanceStatuses[0].InstanceState.Name, nil
}
