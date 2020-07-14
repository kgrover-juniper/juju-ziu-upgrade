#!/bin/bash

juju run-action contrail-agent/0 upgrade &&
juju run-action contrail-agent-csn/0 upgrade
