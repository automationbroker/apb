#!/bin/sh

if [ -z "$1" ]
then
  echo "No namespace supplied. Exiting."
  echo "Usage: ./setup_namespace.sh <namespace>"
  exit
fi

oc new-project $1

oc create serviceaccount apb
oc create rolebinding apb --clusterrole=admin --serviceaccount=$1:apb
