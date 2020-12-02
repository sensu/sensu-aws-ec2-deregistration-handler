package main

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	awsConfig.AwsInstanceID = "i-1234567890abcdef0"
	assert.Error(checkArgs(event))
	awsConfig.AllowedInstanceStates = "running"
	assert.Error(checkArgs(event))
	sensuAPIURL = "http://localhost:8080"
	assert.Error(checkArgs(event))
	sensuAPIKey = "e2bf4da0-ffcc-4744-b29c-94ff9a504e38"
	assert.NoError(checkArgs(event))
}
