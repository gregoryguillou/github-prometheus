# github-prometheus

A Prometheus Exporter for Github that can be customized with GraphQL

## How to install the project?

The `deployment/kubernetes` directory contains an example of a deployment
running on kubernetes. To configure the application:

- download, modify and apply `config-template.yaml`; change the following keys:
  - `app_id` should be the Github application id
  - `client_id` should be the Github client id
  - `client_secret` should be the Github client secret
  - `private_key` should be the Github application private key
- download, modify and apply `deployment.yaml`
