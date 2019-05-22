#!/bin/bash -e

mydir=$(dirname $0)

workspace=default
env_vars="$mydir/env-vars"
###################################
# You should not change code below
improve_and_colorize_tf_dot () {
    sed 's/\[root\] //g' | \
    gvpr -c 'N {style="filled"; if(match($.name,"output.") != -1) {fillcolor="gray90";} else {fillcolor="cyan";}} N[shape=="box"] {style="filled"; if(match($.label,"data.") != -1) {fillcolor="pink";} else {fillcolor="yellow";}} N[shape=="diamond"] {style="filled"; fillcolor="green";}'
}
shopt -s expand_aliases
alias tf=terraform
dotfile=tf-graph.dot
pngfile=tf-graph.png
. $env_vars
tf workspace select "$workspace"
set -x
tf graph >$dotfile
improve_and_colorize_tf_dot <$dotfile >$dotfile.color
dot -Tpng $dotfile.color >$pngfile
