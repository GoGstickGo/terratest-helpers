# Include the root `terragrunt.hcl` configuration. The root configuration contains settings that are common across all
# components and environments, such as how to configure remote state.
locals {
  root_vars   = read_terragrunt_config(find_in_parent_folders("root_vars.hcl"))
}

include "_envcommon" {
  path = "${dirname(find_in_parent_folders())}/_envcommon/iam-policy.hcl"
}

include "root" {
  path = find_in_parent_folders()
}

# ---------------------------------------------------------------------------------------------------------------------
# MODULE PARAMETERS
# These are the variables we have to pass in to use the module. This defines the parameters that are common across all
# environments.
# ---------------------------------------------------------------------------------------------------------------------

inputs = {
  iam_policy_name = "DummyTest-${local.root_vars.locals.aws_region}"
  iam_source_policy_documents = [
    jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "secretsmanager:GetSecretValue"
          ]
          Resource = "arn:aws:secretsmanager:${local.root_vars.locals.aws_region}:${local.root_vars.locals.account_id}:secret:*"
          Effect   = "Allow"
        }
      ]
    })
  ]
}
