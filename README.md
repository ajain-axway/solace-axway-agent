# SOLACE AXWAY AGENT

Axway Agent for provisioning AsyncAPIs into Solace Brokers. 

## Development 
### Prerequisites

* Golang (v 1.16+)
* Make
* Docker and Docker-Compose for integration tests

### Setup Development Environment 

* Solace-Axway-Agent is based on [solace-iot-team/agent-sdk](https://github.com/solace-iot-team/agent-sdk) which is a fork of [Axway/agent-sdk](https://github.com/Axway/agent-sdk) 
  * how to import `agent-sdk` is documented inline in `go.mod`

### Prepare Environment for Integration Tests

* Provide `.env.local` file in `/testing`
  * sample is in `/sample/.env`
* Start Docker-Compose to bring up 
  * Solace-Connector
  * Notifier Service
* Start testing by `make integrationtest`
  * set environment variables (sample is located in `/testing`)

### How to build

* Checkout repository
* Build project
  `make build`
* Linter
  `make lint`

### Code Generation
Solace-Connector and Notifier HTTP-Clients are generated. Detailed information is located in `/specs`
# How to use

## Prerequisites

### Axway Central

* Create Public/Private Key Pair as `PEM`-files
`openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048`
  
* Create Axway Central Service Account
   * Amplify Central `https://central.eu-fr.axway.com` Section `Access`
   * Register Service Account `Add Service Account`
     * Upload / Copy public key PEM 
   
### Solace Environment
* Solace Connector [solace-iot-team/platform-api](https://github.com/solace-iot-team/platform-api)
   * Connector URL
   * Connector Admin username and password
   * Connector Org-Admin username and password
   
For each Axway `Environment` a Solace Connector `Organization` must be provisioned (by convention: same names)  

## Run agent

Configuration of the agent can get provided by a config-file ('solace_axway_agent.yml') or by defining environment variables (still a minimum config-file must be provided, see `sample/sample_min_solace_axway_agent.yml`).


### Prepare `solace_axway_agent.yml` configuration
* Prepare and configure `solace_axway_agent.yml` file. Sample is located in `sample/`
* Or set environment variables. Sample is located in `sample/`
  * Although all configuration options can get defined via environment variables, Solace-Axway-Agent must have access to a minimum `solace_axway_agent.yml` configuration file. This file can get located alongside the executable (same directory) or the directory containing the configuration file can get defined as option `--pathConfig`

### Execute `solace-axway-agent` 
* `./solace-axway-agent --pathConfig /Users/jt/myproject/solace/axway-agent/solace-agent-config`

### Check Health

Health checks (accessibility) of Axway Central and Solace Connector can get accessed via a web service exposed by the agent:

Sample of an agent running on localhost:

* `curl http://localhost:8989/status/central`
* `curl http://localhost:8989/status/solace`


### Environment / Configuration


```yaml
log:
  level: trace
# optional - directory containing SSL/TLS public certificates of endpoints this agent is esablishing connections to
ssl_cert_dir: "/path/to/directory"
# Configuration options offered by Axway Agent SDK 
central:
  # Pollinterval agent is polling Axway Central
  pollInterval: 10s
  # Axway Central API Endpoint
  url: https://central.eu-fr.axway.com
  # Axway Central Organization ID 
  organizationID: 12345
  # Axway Central Environment 
  environment: abc-efg-1
  auth:
    # Axway Central Service Account 
    clientID: DOSA_abc123
    # Path and Filename of Axway Central Service Account private key as PEM 
    privateKey: "/path/to/private_key.pem"
      # Optional - PEM content as one line PEM 
      #            will be written as PEM in central.auth.privateKey file defined in here
      #            can get used to bootstrap and share PEM via environment variable
      # use awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' cert-name.pem  to transform PEM file 
      data: "-----BEGIN PRIVATE KEY----- ..."
    # Path and Filename of Axway Central Service Account public key as PEM
    publicKey: "/path/to/public_key.pem"
      # Optional - PEM content as one line PEM 
      #            will be written as PEM in central.auth.publicKey file defined in here
      #            can get used to bootstrap and share PEM via environment variable
      # use awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' cert-name.pem  to transform PEM file 
      data: "-----BEGIN PUBLIC KEY----- ..."
  # configuration related to management of subscriptions (subscribe / unsubscribe)    
  subscriptions:
    # SMTP Notification options
    notifications:
      # only SMTP supported for Solace-Axway-Agent 
      type:
      - smtp
      smtp:
        # SMTP host 
        host: smtp.host.com 
        # SMTP Port
        port: 1234
        # Sender of notification 
        fromAddress: sender@host.com
        # SMTP authentication
        authType: plain
        # SMTP server username
        username: theusername
        # SMTP server password
        password: thepassword
        subscribe:
          # go-template for subscribe emails
          body: >
            <p>A Subscription just got created for AsyncAPI {{.CatalogItemName}}.</p>Detailed description of the AsyncAPI can get found at: <a href={{.CatalogItemURL}}">Axway Central {{.CatalogItemName}} </a> <p></p><p>The subscribed AsyncAPI is secured with <ul><li>Username: <b> {{.ClientID}} </b></li><li>Password: <b> {{.ClientSecret}} </b></li></ul></p><p>AsyncAPI runtime internal AppId: {{.APIManagerID}} </p>
          # go-template for rendering credentials in subscribe emails
          oauth: "</br></a>Your API is secured using Username: <b> {{.ClientID}} </b> and Password=<b> {{.ClientSecret}} </b>" 
        unsubscribe:
          # go-template for unsubscribe emails
          body: >
            <p>A Subscription just got unsubscribed for AsyncAPI {{.CatalogItemName}}.</p>Detailed description of the AsyncAPI can get found at: <a href={{.CatalogItemURL}}">Axway Central {{.CatalogItemName}} </a> <p>
# Solace-Axway-Agent specific 
bootstrapping:
  # Publish (idempotent) Axway Subscription Schema in Axway Central 
  # Publishing is executed once per Solace-Axway-Agent start-up
  publishSubscriptionSchema: true
  # Assignment of Axway Subscription Schema to AsyncAPIs with attribute solace-webhook-enabled==true
  # see also solaceconst.go 
  processSubscriptionSchema: true
  # Pollinterval in seconds for querying subscriptions
  processSubscriptionSchemaInterval: 60
  
# Solace Connector specific configurations
connector:
  # Solace Connector endpoint
  url: http://url:port/path
  # Solace Connector admin user
  adminUser: admin
  # Solace Connector admin password
  adminPassword: secret
  # Solace Connector organization user 
  orgUser: user
  # Solace Connector organization user password
  orgPassword: secret
  # Enable / disable TLS certificate validation of connector endpoint
  #   provide (root)-certificate of endpoint in ssl_cert_dir to enable it
  acceptinsecurecertificates: false  
  # Enable/ disable logging of HTTP-REST requests sent to Solace-Connector
  #   sensitive information will be written to the log (e.g. usernames and passwords of AsyncAPI subscriptions)
  #   only for debugging and development 
  logBody: true
  # Enable / disable logging of HTTP Headers sent to Solace-Connector
  #   sensitive information will be written to the log (username / password used to authenticate against Solace-Connector)
  #   only for debugging and development 
  logHeader: true

# Notifier Endpoint configuration
notifier:
  # Enable / disable notifier
  enabled: false
  # Notifier endpoint
  url: http://notifier.endpoint.com/path
  # Enable / disable TLS certificate validation of notifier endpoint
  #   provide (root)-certificate of endpoint in ssl_cert_dir to enable it
  acceptinsecurecertificates: false
  # Authentication type
  #   basic: BasicAuth
  #   header: HTTPHeader / Value
  apiAuthType: basic
  # Authentication Type
  #   basic: username
  #   header: headername
  apiConsumerKey: username/headername
  # Authentication Type
  #   basic: password
  #   header: headervalue apikey, etc. 
  apiConsumerSecret: abcefg
  # Health message sent to Notifier health endpoint 
  healthmessage: "ping"
```