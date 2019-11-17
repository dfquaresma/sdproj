#!/bin/bash
date
set -x

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "RESULTS_PATH: ${RESULTS_PATH:=./results}"
echo "NUMBER_OF_EXPERIMENTS: ${NUMBER_OF_EXPERIMENTS:=10}"

mkdir -p ${RESULTS_PATH}
for expid in `seq 1 ${NUMBER_OF_EXPERIMENTS}`;
do
    # to tear down old containeres
    sudo docker ps --filter name=listfiller* -aq | sudo xargs docker stop | sudo xargs docker rm
    EXPID=${expid} RESULTS_PATH=${RESULTS_PATH} bash ./seq-workload.sh
done