#!/bin/bash

export OMA_DEPLOY_PWD=`pwd`
export OMA_DEPLOY_MODEL=`pwd`/../../cmd/model/model
export OMA_DEPLOY_ROOT=$1

exec bash --rcfile devel-rcfile.sh
