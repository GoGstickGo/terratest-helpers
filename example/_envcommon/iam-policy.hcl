locals {
  # Expose the base source URL so different versions of the module can be deployed in different environments.
  base_source_url = "git::git@github.com/cloudposse/terraform-aws-iam-policy.git//."
}

terraform {
  source = "${local.base_source_url}?ref=v1.0.1"
}

prevent_destroy = false

# ---------------------------------------------------------------------------------------------------------------------
# MODULE PARAMETERS
# ---------------------------------------------------------------------------------------------------------------------
inputs = {
  iam_policy_enabled = true
}