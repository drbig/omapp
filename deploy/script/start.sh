#!/bin/bash

ROOT=/home/vagrant/omapp

export OMA_DATA_ROOT=$ROOT
export OMA_WEB_SECRET=dupa.9
export OMA_WEB_ORIGIN=http://192.168.50.51:8080
export OMA_DB_DRIVER=sqlite3
export OMA_DB_ARGS=${ROOT}/db.bin
export OMA_DB_VERBOSE=0
export OMA_B_MOUNT=127.0.0.1:7000
export OMA_U_MOUNT=127.0.0.1:8000
export OMA_Q_MOUNT=127.0.0.1:11300

OMA_DEPLOY_CMDS=(worker uploader backend)
for CMD in ${OMA_DEPLOY_CMDS[*]}; do
	${ROOT}/bin/${CMD} > ${ROOT}/logs/${CMD}.log 2>&1 &
	echo $! > ${ROOT}/run/${CMD}.pid
done
