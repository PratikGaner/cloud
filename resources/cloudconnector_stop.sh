#!/bin/sh

#
# Copyright (c) 2022 Contributors to the Eclipse Foundation
#
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Apache License 2.0 which is available at
# https://www.apache.org/licenses/LICENSE-2.0
#
# SPDX-License-Identifier: Apache-2.0
#

TIMEOUT=10
cnt=0
pkill -TERM cloudconnector
# while [ -n "`pidof cloudconnecor`" ]; do # not working well on raspberry
while [ -n "`ps -ef | grep 'cloudconnector ' | grep -v grep`" ]; do
  cnt=`expr $cnt + 1`
#   cnt=$(($cnt + 1)) # bash alternative if expr not available
  if [ $cnt -gt $TIMEOUT ]; then
    echo "### TIMEOUT waiting for cloudconnector stop ($TIMEOUT s). Killing instance [`pidof cloudconnector`]!"
    pkill -9 cloudconnecor
    exit 1
  fi
  sleep 1
done

echo "cloudconnector stopped in ($cnt s)."
exit 0
