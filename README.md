# gitlab-machine

A tool to deploy on demand Gitlab CI/CD runners using the Custom Executor.

## What is it?

The Custom Executor is a mechanism from GitLab CI/CD that allows to plug specific runners to the CI/CD pipeline. This tool is a wrapper around the Custom Executor that allows to deploy on demand virtual instances.

It is currently in pre-alpha state, with support only for VMware Cloud Director and Windows machines (using SSH communication).

It should probably be replace with https://gitlab.com/gitlab-org/ci-cd/custom-executor-drivers/autoscaler

## Expected config

```yaml
# The config file is a YAML file with the following structure:
log_level: debug

vcd:
  motd: "Deploying a dedicated VM using https://github.com/juanfont/gitlab-machine"
  url: https://cloudirector
  org: tenant
  vdc: virtualdatancer
  insecure: false
  user: username
  password: password
  vdc_network: orgvdcnetwrok
  catalog: vcdcatalogue
  template: Windows_10
  num_cpus: 8
  cores_per_socket: 8
  memory_mb: 8192
  storage_profile: storageprofile
  default_password: VMpassword
```

## More info

- [GitLab Custom Executor](https://docs.gitlab.com/runner/executors/custom.html)
- [Windows SaaS runners](https://docs.gitlab.com/ee/ci/runners/saas/windows_saas_runner.html)
- [A practical guide to the Custom Executor](https://medium.com/ci-t/a-practical-guide-to-gitlab-runner-custom-executor-drivers-bc6e6562647c)
