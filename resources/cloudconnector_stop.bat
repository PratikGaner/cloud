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

rem kill with /F option, otherwise the process could not be terminated
taskkill /F /im cloudconnector.exe
