# ---------------------------------------------------------------------------------------------------------------------
# TERRAGRUNT CONFIGURATION
# Terragrunt is a thin wrapper for Terraform that provides extra tools for working with multiple Terraform modules,
# remote state, and locking: https://github.com/gruntwork-io/terragrunt
# ---------------------------------------------------------------------------------------------------------------------
locals {
  # Automatically load root-level variables
  root_vars = read_terragrunt_config(find_in_parent_folders("root_vars.hcl"), { locals = {} })
  #extra_vars = read_terragrunt_config(find_in_parent_folders("app.hcl", "infra.hcl"), { locals = {} })
  mandatory_tags = read_terragrunt_config(find_in_parent_folders("mandatory_tags.hcl"), { locals = { mandatory_tags = {} } })
  child_tags     = read_terragrunt_config(find_in_parent_folders("child_tags.hcl"), { locals = { child_tags = {} } })
  merged_tags    = merge(local.mandatory_tags.locals.mandatory_tags, local.child_tags.locals.child_tags)
  #populate tags map
  tags_map = { locals = { tags = local.merged_tags } }

  # Extract the variables we need for easy access
  account_name = local.root_vars.locals.stage
  aws_profile  = local.root_vars.locals.stage
  account_id   = local.root_vars.locals.account_id
  aws_region   = local.root_vars.locals.aws_region
  region       = local.root_vars.locals.aws_region
  environment  = local.root_vars.locals.environment
  tenant       = local.root_vars.locals.tenant
}

# Generate an AWS provider block
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<-EOF
    provider "aws" {
      region = "${local.aws_region}"
      # profile = "${local.aws_profile}"
      # Only these AWS Account IDs may be operated on by this template
      allowed_account_ids = ["${local.account_id}"]
    }
  EOF
}

generate "versions" {
  path      = "versions.tf"
  if_exists = "skip"
  contents  = <<EOF
     terraform { 
       required_version = ">= 1.4.1"
       required_providers {
         aws = {
           source  = "hashicorp/aws"
           version = ">= 4.59.0"
         }
         local = {
           source  = "hashicorp/local"
           version = ">= 2.1.0"
         }
         null = {
           source  = "hashicorp/null"
           version = ">= 3.1.1"
         }
       }
     }
   EOF
}

//local development
remote_state {
  backend = "local"
  config = { path = "${get_parent_terragrunt_dir()}/${path_relative_to_include()}/terraform.tfstate" }

  generate = {
  path = "backend.tf"
    if_exists = "overwrite"
  }
}


inputs = merge(
  local.root_vars.locals,
  local.tags_map.locals
)

terraform {
  after_hook "after_delete_terragrunt_cache" {
    commands     = ["validate", "apply", "destroy"]
    execute      = ["rm", "-rf", ".terragrunt-cache"]
    working_dir  = "${get_terragrunt_dir()}"
    run_on_error = true
  }

  after_hook "after_delete_terraform_lock" {
    commands     = ["validate", "apply", "destroy"]
    execute      = ["rm", ".terraform.lock.hcl"]
    working_dir  = "${get_terragrunt_dir()}"
    run_on_error = true
  }
}