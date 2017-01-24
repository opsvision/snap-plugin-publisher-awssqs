#!/usr/bin/env bash

# This is a helper script for loading our plugin to expediate testing during development

PLUGIN='snap-plugin-publisher-awssqs'
TASK='~/tasks/awssqs.yaml'

if [ -e ${GOPATH} ]; then echo "GOPATH may not be set correctly"; echo -1; fi

if [ -e ${GOPATH}/bin/${PLUGIN} ]; then rm ${GOPATH}/bin/${PLUGIN}; fi

go install && \
  snaptel plugin load ~/golang/bin/${PLUGIN} && \
  snaptel task create -t ${TASK} && \
  snaptel task list
