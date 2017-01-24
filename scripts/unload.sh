#!/usr/bin/env bash

# This is a helper script for unloading our plugin to expedite testing during development

snaptel task list | tail -1 | awk '{ print $1 }' | xargs snaptel task stop

snaptel task list | tail -1 | awk '{ print $1 }' | xargs snaptel task remove

snaptel plugin unload publisher awssqs 1

if [ -e /tmp/awssqs-debug.log ];then sudo rm -f /tmp/awssqs-debug.log; fi
