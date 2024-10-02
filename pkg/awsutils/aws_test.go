package awsutils_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoGstickGo/terratest-helpers/pkg/awsutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/stretchr/testify/assert"
)

type MockAWSConfigLoader struct {
	LoadConfigFunc func(ctx context.Context, region string) (aws.Config, error)
}

// LoadConfig calls the mock function, allowing the user to define the behavior.
func (m *MockAWSConfigLoader) LoadConfig(ctx context.Context, region string) (aws.Config, error) {
	return m.LoadConfigFunc(ctx, region)
}

type MockWorkMailClient struct {
	DeleteOrganizationFunc func(ctx context.Context, params *workmail.DeleteOrganizationInput, optFns ...func(*workmail.Options)) (*workmail.DeleteOrganizationOutput, error)
}

type MockEC2Client struct {
	DescribeNetworkInterfacesFunc func(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, opts ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DeleteNetworkInterfaceFunc    func(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput, opts ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error)
}

func (m *MockWorkMailClient) DeleteOrganization(ctx context.Context, params *workmail.DeleteOrganizationInput, optFns ...func(*workmail.Options)) (*workmail.DeleteOrganizationOutput, error) {
	return m.DeleteOrganizationFunc(ctx, params, optFns...)
}

func (m *MockEC2Client) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, opts ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return m.DescribeNetworkInterfacesFunc(ctx, params, opts...)
}

func (m *MockEC2Client) DeleteNetworkInterface(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput, opts ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error) {
	return m.DeleteNetworkInterfaceFunc(ctx, params, opts...)
}

// Unit test for DeleteWorkMailOrganization.
func TestMockDeleteWorkMailOrganization(t *testing.T) {
	t.Parallel()
	// Mock the AWS configuration loader.
	mockLoader := &MockAWSConfigLoader{
		LoadConfigFunc: func(_ context.Context, _ string) (aws.Config, error) {
			// Return a dummy AWS config here.
			return aws.Config{}, nil
		},
	}

	// Call the LoadConfig method on the mock loader.
	cfg, err := mockLoader.LoadConfig(context.TODO(), "us-east-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert the returned config (you can further assert the values if needed).
	if cfg.Region != "" {
		t.Errorf("Expected empty region, got %s", cfg.Region)
	}

	mockClient := &MockWorkMailClient{
		DeleteOrganizationFunc: func(ctx context.Context, params *workmail.DeleteOrganizationInput, optFns ...func(*workmail.Options)) (*workmail.DeleteOrganizationOutput, error) {
			// Simulate success.
			return &workmail.DeleteOrganizationOutput{}, nil
		},
	}

	// Call the function under test.
	err = awsutils.DeleteWorkMailOrganization(t, "test-org-id", mockClient)

	// Assert no error.
	assert.NoError(t, err)
}

func TestMockFailureDeleteWorkMailOrganization(t *testing.T) {
	t.Parallel()
	// Mock the AWS configuration loader.
	mockLoader := &MockAWSConfigLoader{
		LoadConfigFunc: func(_ context.Context, _ string) (aws.Config, error) {
			// Return a dummy AWS config here.
			return aws.Config{}, nil
		},
	}

	// Call the LoadConfig method on the mock loader.
	cfg, err := mockLoader.LoadConfig(context.TODO(), "us-east-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert the returned config (you can further assert the values if needed).
	if cfg.Region != "" {
		t.Errorf("Expected empty region, got %s", cfg.Region)
	}

	mockClient := &MockWorkMailClient{
		DeleteOrganizationFunc: func(ctx context.Context, params *workmail.DeleteOrganizationInput, optFns ...func(*workmail.Options)) (*workmail.DeleteOrganizationOutput, error) {
			return nil, fmt.Errorf("failed success")
		},
	}

	// Call the function under test.
	err = awsutils.DeleteWorkMailOrganization(t, "test-org-id", mockClient)

	// Assert  error.
	assert.ErrorContainsf(t, err, "failed to delete WorkMail organization", err.Error())
}

// TestRemoveENI tests the RemoveENI function.
func TestMockRemoveENI(t *testing.T) {
	t.Parallel()
	// Mock the AWS configuration loader.
	mockLoader := &MockAWSConfigLoader{
		LoadConfigFunc: func(_ context.Context, _ string) (aws.Config, error) {
			// Return a dummy AWS config here.
			return aws.Config{}, nil
		},
	}

	// Call the LoadConfig method on the mock loader.
	cfg, err := mockLoader.LoadConfig(context.TODO(), "us-east-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert the returned config (you can further assert the values if needed).
	if cfg.Region != "" {
		t.Errorf("Expected empty region, got %s", cfg.Region)
	}
	mockClient := &MockEC2Client{
		DescribeNetworkInterfacesFunc: func(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, opts ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
			return &ec2.DescribeNetworkInterfacesOutput{
				NetworkInterfaces: []types.NetworkInterface{
					{
						NetworkInterfaceId: aws.String("eni-12345678"),
						VpcId:              aws.String("vpc-123456"),
						Attachment:         nil, // Simulating an unused ENI.
					},
				},
			}, nil
		},
		DeleteNetworkInterfaceFunc: func(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput, opts ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error) {
			return &ec2.DeleteNetworkInterfaceOutput{}, nil
		},
	}

	counter, err := awsutils.RemoveENI(t, "vpc-123456", mockClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if counter != 1 {
		t.Fatalf("expected to delete 1 ENI, got %d", counter)
	}
}

func TestMockFailiureRemoveENI(t *testing.T) {
	t.Parallel()
	// Mock the AWS configuration loader.
	mockLoader := &MockAWSConfigLoader{
		LoadConfigFunc: func(_ context.Context, region string) (aws.Config, error) {
			// Return a dummy AWS config here.
			return aws.Config{}, nil
		},
	}

	// Call the LoadConfig method on the mock loader.
	cfg, err := mockLoader.LoadConfig(context.TODO(), "us-east-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assert the returned config (you can further assert the values if needed).
	if cfg.Region != "" {
		t.Errorf("Expected empty region, got %s", cfg.Region)
	}

	mockClient := &MockEC2Client{
		DescribeNetworkInterfacesFunc: func(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, opts ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error) {
			return nil, fmt.Errorf("failed success")
		},
		DeleteNetworkInterfaceFunc: func(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput, opts ...func(*ec2.Options)) (*ec2.DeleteNetworkInterfaceOutput, error) {
			return &ec2.DeleteNetworkInterfaceOutput{}, nil
		},
	}

	counter, err := awsutils.RemoveENI(t, "vpc-123456", mockClient)
	if err == nil {
		t.Fatalf("expected error, got %v", err)
	}
	if counter != 0 {
		t.Fatalf("expected to delete 0 ENI, got %d", counter)
	}
}
