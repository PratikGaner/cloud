@echo off
rem
rem Copyright (c) 2022 Contributors to the Eclipse Foundation
rem
rem See the NOTICE file(s) distributed with this work for additional
rem information regarding copyright ownership.
rem
rem This program and the accompanying materials are made available under the
rem terms of the Apache License 2.0 which is available at
rem https://www.apache.org/licenses/LICENSE-2.0
rem
rem SPDX-License-Identifier: Apache-2.0
rem

setlocal

rem Check if cloud connector is already started
FOR /F %%x IN ('tasklist /NH /FI "IMAGENAME eq cloudconnector.exe"') DO IF %%x == cloudconnector.exe echo Cloud Connector already started, exiting^! & timeout 2 & EXIT 1

rem Configuration file in json format with flags values, configure with parameter -configFile.
if defined CLOUD_CONNECTOR_CONFIG set "ARGUMENTS=%ARGUMENTS% -configFile=%CLOUD_CONNECTOR_CONFIG%"

rem Connection string for the Azure IoT Hub connectivity, configure with parameter -connectionString.
if defined CONNECTION_STRING set "ARGUMENTS=%ARGUMENTS% -connectionString=%CONNECTIONS_STRING%"

rem ID Scope for the Azure DPS authentication, configure with parameter -idScope.
if defined ID_SCOPE set "ARGUMENTS=%ARGUMENTS% -idScope=%ID_SCOPE%"

rem The file for the message mappings configuration, configure with parameter -messageMapperConfig ("message-mapper-config.json").
if defined MESSAGE_MAPPER_CONFIG set "ARGUMENTS=%ARGUMENTS% -messageMapperConfig=%MESSAGE_MAPPER_CONFIG%"

rem List of passthrough device topics, configure with parameter -passthroughDeviceTopics.
if defined PASSTHROUGH_DEVICE_TOPICS set "ARGUMENTS=%ARGUMENTS% -passthroughDeviceTopics=%PASSTHROUGH_DEVICE_TOPICS%"

rem List of passthrough command names, configure with parameter -passthroughCommandNames.
if defined PASSTHROUGH_COMMAND_NAMES set "ARGUMENTS=%ARGUMENTS% -passthroughCommandNames=%PASSTHROUGH_COMMAND_NAMES%"

rem User-specified tenant id, configure with parameter -tenantId (default "defaultTenant").
if defined TENANT_ID set "ARGUMENTS=%ARGUMENTS% -tenantId=%TENANT_ID%"

rem Address of the local MQTT broker, configure with parameter -localAddress (default "tcp://localhost:1883").
if defined LOCAL_ADDRESS set "ARGUMENTS=%ARGUMENTS% -localAddress=%LOCAL_ADDRESS%"

rem Username for authentication to the local client, configure with parameter -localUsername.
if defined LOCAL_USERNAME set "ARGUMENTS=%ARGUMENTS% -localUsername=%LOCAL_USERNAME%"

rem Password for authentication to the local client, configure with parameter -localPassword.
if defined LOCAL_PASSWORD set "ARGUMENTS=%ARGUMENTS% -localPassword=%LOCAL_PASSWORD%"

rem Path to Hub certificate, configure with parameter -caCert (default "iothub.crt").
if defined CA_CERT_PATH set "ARGUMENTS=%ARGUMENTS% -caCert=%CA_CERT_PATH%"

rem Log file location, configure with parameter -logFile (default "logs/log.txt")
if defined LOG_FILE set "ARGUMENTS=%ARGUMENTS% -logFile=%LOG_FILE%"

rem Log level, configure with parameter -logLevel.
rem Possible values: ERROR, WARN, INFO, DEBUG, TRACE (default "INFO").
if defined LOG_LEVEL set "ARGUMENTS=%ARGUMENTS% -logLevel=%LOG_LEVEL%"

rem Log file size in MB before it gets rotated, configure with parameter -logFileSize (default 2).
if defined LOG_FILE_SIZE set "ARGUMENTS=%ARGUMENTS% -logFileSize=%LOG_FILE_SIZE%"

rem Log file max rotations count, configure with parameter -logFileCount (default 5).
if defined LOG_FILE_COUNT set "ARGUMENTS=%ARGUMENTS% -logFileCount=%LOG_FILE_COUNT%"

rem Log file rotations max age in days, configure with parameter -logFileMaxAge (default 28).
if defined LOG_FILE_MAX_AGE set "ARGUMENTS=%ARGUMENTS% -logFileMaxAge=%LOG_FILE_MAX_AGE%"

echo %cd%\cloudconnector.exe %ARGUMENTS%
start cloudconnector.exe %ARGUMENTS%
