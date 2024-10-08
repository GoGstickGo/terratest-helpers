package awsutils

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoGstickGo/terratest-helpers/pkg/parameters"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/gruntwork-io/terratest/modules/logger"
)

// AWS config.
type DefaultAWSConfigLoader struct{}

// LoadConfig loads the AWS configuration using the AWS SDK.
func (d *DefaultAWSConfigLoader) LoadConfig(ctx context.Context, region string) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, config.WithRegion(region))
}

type WorkMailClient interface {
	DeleteOrganization(ctx context.Context, params *workmail.DeleteOrganizationInput, optFns ...func(*workmail.Options)) (*workmail.DeleteOrganizationOutput, error)
}

type EC2Client interface {
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, opts ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DeleteNetworkInterface(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput, opts ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error)
}

func LoadEC2Client(region string) (*ec2.Client, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Return the EC2 client
	return ec2.NewFromConfig(cfg), nil
}

func RemoveENI(t *testing.T, vpcID string, svc EC2Client) (int32, error) {
	var counter int32

	logger.Log(t, "Remove unused ENIs in VPC Id:", vpcID)

	describeInput := &ec2.DescribeNetworkInterfacesInput{}
	result, err := svc.DescribeNetworkInterfaces(context.TODO(), describeInput)
	if err != nil {
		return 0, fmt.Errorf("error describing network interfaces: %w", err)
	}

	// Delete each unused network interface
	for _, networkInterface := range result.NetworkInterfaces {
		if networkInterface.Attachment == nil || networkInterface.Attachment.InstanceId == nil && networkInterface.VpcId == &vpcID {
			// ENI is not attached to any instance, safe to delete
			deleteInput := &ec2.DeleteNetworkInterfaceInput{
				NetworkInterfaceId: networkInterface.NetworkInterfaceId,
			}

			_, deleteErr := svc.DeleteNetworkInterface(context.TODO(), deleteInput)
			if deleteErr != nil {
				logger.Log(t, "Error deleting ENI", deleteErr) // Log the error for debugging purposes("error deleting ENI: %v", deleteErr).
			} else {
				counter++
			}
		}
	}

	logger.Log(t, "Number of ENIs deleted:", counter)

	return counter, nil
}

func DeleteWorkMailOrganization(t *testing.T, orgID string, client WorkMailClient) error {
	// Load the default AWS configuration.

	logger.Log(t, "Remove Wokrmail ORGId:", orgID)

	loader := &DefaultAWSConfigLoader{}
	_, err := loader.LoadConfig(context.TODO(), parameters.AWSRegion)

	if err != nil {
		return fmt.Errorf("AWS Auth error %w", err)
	}

	// Create the input parameters for the DeleteOrganization API call.
	input := &workmail.DeleteOrganizationInput{
		OrganizationId: aws.String(orgID),
	}

	// Call the DeleteOrganization API.
	_, err = client.DeleteOrganization(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to delete WorkMail organization: %w", err)
	}

	return nil
}
