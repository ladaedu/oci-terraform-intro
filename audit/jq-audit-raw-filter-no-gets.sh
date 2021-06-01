#! /bin/bash -eu

jq -r '{data: [.data[]|select(.data.request.action != "GET" and .data.identity.principalName != null)]}' $@
