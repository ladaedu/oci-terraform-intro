#! /bin/bash -eu

jq -r '.data[]|"\(.eventTime)\t\(.data.identity.principalName)\t\(.source)\t\(.data.eventName)\t\(.data.resourceName)\t\(.data.request.action)\t\(.data.response.status)"' $@
