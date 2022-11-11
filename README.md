# Cloud Connector

The repository contains the cloud connector component used in the Eclipse Leda stack.
# Build
## Build prerequisites
- Go 1.17 (or higher) needs to be available on the build host.
- Set the `GOOS` & `GOARCH` environment variables to build the cloud connector for a specific platform.
## Supported Go platforms
The build process supports the following Golang platforms:
- linux/arm
- linux/arm64
- linux/amd64
- windows/amd64
## Performing a build
### Manual build
Navigate to `<root_level>/cmd` where <root_level> is the root of the repo and execute:<br>
`go build -ldflags="-s -w -X "main.version=$CLOUD_CONNECTOR_VERSION"" -trimpath -mod=readonly -o target/cloudconnector` or <br>
`go build -ldflags="-s -w -X "main.version=$CLOUD_CONNECTOR_VERSION"" -trimpath -mod=readonly -o target/cloudconnector.exe` for the windows/amd64 platform<br>
where `$CLOUD_CONNECTOR_VERSION` has to be replaced with a required, custom version of the cloud connector.<br>

The build process produces a binary with name `cloudconnector` or `cloudconnector.exe`, located in `target` subfolder.

### By executing a shell script
Navigate to `<root_level>/cmd/build.sh` where <root_level> is the root of the repo and execute:<br>
`sudo .\build.sh <version>` where `<version>` and `things` are optional.<br>
If provided, `<version>` has to be replaced with a required, custom version of the cloud connector.<br>

The build process produces a separate tarball archive per platform, located in `target` subfolder and additionally an aggregated assembly archive that packs all platform-specific assemblies.<br>
The naming convention of a platform assembly is `cloud_connector-$GOOS-$GOARCH.tar.gz`, where `$GOOS` and `$GOARCH` represents respectively the Golang platform OS and architecture.<br>
Each generated assembly packs the cloud connector binary, the root CA certificate, the cloud connector start/stop scripts for the specific platform and optionally the required static resources and configuration files.   

# Run
## Supported options to get device connection info
There are several options how the cloud connector obtains its connection info:
- Device connection string is provided in cloud connector config (either with SharedAccessKey or device key+certificate) – in this case cloud connector directly uses the provided data and tries to connect to Azure IoT Hub. Device should be already registered in Azure IoT Hub.
- Configuration includes device key+certificate and idScope parameter – in this case the cloud connector connects to Azure DPS in order to register the device to an IoT Hub. Device should be pre-enrolled in Azure DPS.

## Configuration parameters
The cloud connector supports the following configuration parameters:

- Connection string
Optional. Represents the connection string, used for the connectivity to the Azure IoT hub device.<br>
The name of the parameter is `connectionString`, when passed as a flag to the binary, or `CONNECTION_STRING`, when preset as an environment variable.

- Device Certificate Path
Optional. Represents the path to the X.509 certificate, used for the device authentication.<br>
The name of the parameter is `cert`, when passed as a flag to the binary, or `CERT_PATH`, when preset as an environment variable.

- Device Certificate Key Path
Optional. Represents the path to the private key, used for the device authentication.<br>
The name of the parameter is `key`, when passed as a flag to the binary, or `KEY_PATH`, when preset as an environment variable.

- ID Scope
Optional with default value ``. Represents the ID Scope for the Azure DPS authentication.<br>
The name of the parameter is `idScope`, when passed as a flag to the binary, or `ID_SCOPE`, when preset as an environment variable.

- Tenant ID
Optional with default value `defaultTenant`. Represents the tenant ID, which is used only internally with components that "speak" Ditto protocol, e.g. container management component.<br>
The name of the parameter is `tenantId`, when passed as a flag to the binary, or `TENANT_ID`, when preset as an environment variable.

- Passthrough Device Topics
Optional. Represents the list of the local MQTT topics that should be passed from the local MQTT broker to the cloud directly, without any topic remappings or payload transformation.<br>
The name of the parameter is `passthroughDeviceTopics`, when passed as a flag to the binary, or `PASSTHROUGH_DEVICE_TOPICS`, when preset as an environment variable.

- Allowed Cloud Message Types
Optional. Represents the list of the cloud command names that should be forwarded from the cloud to the local MQTT broker, without any payload transformations. The topic where the message is published in constructed from the other properties inside the cloud message: `$appId/$cmdName`.<br>
The name of the parameter is `passthroughCommandNames`, when passed as a flag to the binary, or `PASSTHROUGH_COMMAND_NAMES`, when preset as an environment variable.

- Local Address
Optional with default value `tcp://localhost:1883`. Represents the address of the local MQTT broker.<br>
The name of the parameter is `localAddress`, when passed as a flag to the binary, or `LOCAL_ADDRESS`, when preset as an environment variable.

- SAS Token Validity
Optional with default value `1h`. Represents the validity period of the SAS token for device authentication.<br>
The name of the parameter is `sasTokenValidity`, when passed as a flag to the binary, or `SAS_TOKEN_VALIDITY`, when preset as an environment variable.

- Root Certificate Path
Optional with default value `iothub.crt`. Represents the path to the X.509 certificate, used for the connection to the Azure IoT Hub via MQTT.<br>
The name of the parameter is `caCert`, when passed as a flag to the binary, or `CA_CERT_PATH`, when preset as an environment variable.

- Log File Location
Optional with default value `logs/log.txt`. Represents the log file location.<br>
The name of the parameter is `logFile`, when passed as a flag to the binary, or `LOG_FILE`, when preset as an environment variable.

- Message Mapper File Location
Optional with default value `message-mapper-config.json`. Represents the message mappings configuration file location.<br>
The name of the parameter is `messageMapperConfig`, when passed as a flag to the binary, or `MESSAGE_MAPPER_CONFIG`, when preset as an environment variable.

- Config File Location
Optional with default empty value. Represents the connector configuration json file location.<br>
The name of the parameter is `configFile`, when passed as a flag to the binary, or `CLOUD_CONNECTOR_CONFIG`, when preset as an environment variable.

- Log Level
Optional with default value `INFO`. Represents the log level. The possible values are : `TRACE, DEBUG, INFO, ERROR`.<br>
The name of the parameter is `logLevel`, when passed as a flag to the binary, or `LOG_LEVEL`, when preset as an environment variable.

## Authentication mode
The cloud connector authenticates to cloud, using two alternative mechanisms - by SAS token or X.509 certificate.<br>
### SAS token-based device authentication
To connect to the cloud, using SAS token-based device authentication, the cloud connector must be started with provided a
required Azure device connection string in the format 
`HostName=<hostName>;DeviceId=<deviceId>;SharedAccessKey=<base64-encoded-shared-access-key>`.<br>

Optionally, a SAS token validity parameter can be provided to overwrite the default and apply custom SAS token validity period.
### X.509 certificate-based authentication
To connect to the cloud, using X.509 certificate-based authentication, the cloud connector must be started with provided a required device certificate and device certificate key.<br>

Optionally, a connection string in the format `HostName=<hostName>;DeviceId=<deviceId>` can be provided that contains the
device connection info. If omitted, the cloud connector contacts the global Azure Provisioning Service to obtain the info for a given `idScope`.<br>

## Instructions
The cloud connector can be started in both ways:
- As a standalone executable

Run from the commandline `cloudconnector <params>`, where <params> denotes a list of supported cloud connector parameters (listed above), depending on the chosen device authentication mode.<br>
- By executing a shell script

Execute the shell script `cloudconnector_start.sh` (or .bat for the windows/amd64 platform), located at `<root_level>/resources`.<br>
It is required that the script and the binary has to be placed in the same folder.<br>
The configuration parameters shall be passed as parameters to the script or preset as environment variables on the host.

`Note: If the cloud connector is started without specifying a custom X.509 certificate file, the default one, located at <root_level>/resources/iothub.crt needs be copied and placed in the same folder as the cloud connector binary.`

# Contributing
If you want to contribute bug reports or feature requests, please use *GitHub Issues*.

# License and Copyright
This program and the accompanying materials are made available under the
terms of the Apache License 2.0 which is available at
https://www.apache.org/licenses/LICENSE-2.0
For details, please see our license [LICENSE](LICENSE)