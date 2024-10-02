package parameters

var (
	VPCId        = "vpc-01111111111111111"
	AWSRegion    = "us-east-1"
	AWSAccountID = "111111111111"
	TGRootVars   = `
	locals {
		account_id  = "111111111111"
		aws_region  = "us-east-10"
		environment = "test"
		namespace   = "none"
		stage       = "aws-account"
		tenant      = "tt"
		label_order = ["environment", "tenant", "name", "attributes"]
	}
		`
)
