# Deadman-listener

Deadman-listener is a simple utility designed for effectively managing deadman alerts and overseeing the alert pipeline involving Prometheus and Alertmanager. The workflow seamlessly integrates Prometheus (utilizing watchdog prom rules), Alertmanager, and deadman-listener. If deadman-listener fails to receive alerts from Alertmanager, it triggers an alert back to Alertmanager which you can route it to various communication channels, including Slack, Telegram, etc., through the Alertmanager routes and receiver config.

This tools is built based on **[gouthamve/deadman repository](https://github.com/gouthamve/deadman)** with additional enhancements and updates.

## Installation:

You can easily build and deploy deadman-listener using the following options:

- Build binary apps: `make build`
- Build Docker image: `make docker`
- Get help: `./deadman-listener -h`

## Usage:

1. **Alert Generation in Prometheus:**
   
  To continuously generate alerts in Prometheus, add the following rule to the Prometheus configuration:

  ```yaml
  - alert: Watchdog
    expr: vector(1)
    labels:
      severity: deadman
    annotations:
      description: This is a DeadMansSwitch meant to ensure that the entire Alerting pipeline is functional.
  ```

2. **Configuration in Alertmanager:**

  In the Alertmanager cluster configuration, add a route to send webhook notifications to the deployed Deadman process:

  ```yaml
  ...
  routes:
    - receiver: deadman-listener
      group_wait: 0s
      group_interval: 0s
      repeat_interval: 15s
      match:
        severity: deadman
  ...

  receivers:
    - name: deadman-listener
      webhook_configs:
        - url: http://deadman-ip:9095/ping
  ...
  ```

3. **Run Deadman**
  - Run deadman-listener binary
  ```
  ./deadman-listener
  ```

  - Run deadman-listener as alertmanager's sidecar
  ```
  apiVersion: monitoring.coreos.com/v1
  kind: Alertmanager
  metadata:
    name: my-alertmanager
  spec:
    ...
    containers:
    - image: 'docker.io/jatiman/deadman-listener:latest'
      name: deadman-listener
    ...
    
  ```
## Contribution:

Feel free to contribute to deadman-listener by providing feedback, reporting issues, or submitting pull requests. Let's collaborate to enhance and optimize the capabilities of this alert management utility.