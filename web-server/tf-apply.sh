#!/bin/bash -e

mydir=$(dirname "$0")

# This script will run apply, and then will update ~/.ssh/config for Hosts bastion-cluster and bastion-service
# Configure these 5 variables to suit your needs:
bastion_host_in_config=bastion
ssh_config=~/.ssh/config
workspace=default
env_vars="$mydir/env-vars"
###################################
# You should not change code below
shopt -s expand_aliases
alias tf=terraform
. $env_vars
tf workspace select "$workspace"
time tf apply -auto-approve
tf_out=$(tf output --json)
bastion_ip=$(echo "$tf_out" | jq -r '.BastionPublicIP.value[0]' | grep -o "[0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+")
awk -i inplace -v INPLACE_SUFFIX=.bak "
    /^Host / {x=0} 
    /^Host $bastion_host_in_config\$/ {x=\"bastion\"} 
    x==\"bastion\" &&/Hostname/ {gsub(/Hostname.*/,\"Hostname $bastion_ip\");x=0} 
    {print}
    " "$ssh_config"



