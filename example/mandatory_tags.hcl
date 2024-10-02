locals {
  mandatory_tags = {
    #secops
    dept_code       = 000
    dept_name       = "dummy"
    department_name = "dummy"
    #netops
    team_owner = "dummy"
    #pie
    managed_by   = "terraform"
    source       = "github.com/GoGstickGo/terratest-helpers"
    project_name = "dummy"
  }
}
