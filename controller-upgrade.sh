#!/bin/bash

juju run-action --wait contrail-controller/leader upgrade-ziu &&
juju config contrail-analytics image-tag=$1 &&
juju config contrail-analyticsdb image-tag=$1 &&
juju config contrail-agent image-tag=$1 &&
juju config contrail-agent-csn image-tag=$1 &&
juju config contrail-openstack image-tag=$1 &&
juju config contrail-controller image-tag=$1
