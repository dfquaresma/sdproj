#!/bin/bash
date
set -x

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "REQS: ${REQS:=10000}"
echo "TARGET: ${TARGET:=localhost:8080/function/}"
echo "FUNC: ${FUNC:=listfiller}"
echo "FLAGS: ${FLAGS:=gci nogci}"

# To avoid execution without passing environment variables
if [[ (-z "$EXPID") || (-z "$RESULTS_PATH") ]];
then
  echo -e "${RED}EXPID OR RESULTS_PATH MISSING: seq-workload.sh${NC}"
  exit
fi

for flag in ${FLAGS};
do
    FILE_NAME="${RESULTS_PATH}${FUNC}-${flag}${EXPID}.csv"
    echo -e "status;latency" > ${FILE_NAME}
    for i in `seq 1 ${REQS}`
    do
        curl -X GET -o /dev/null -s -w '%{http_code};%{time_total}\n' "${TARGET}${FUNC}-${flag}" >> ${FILE_NAME}
    done
    sed -i 's/,/./g' ${FILE_NAME}
    sed -i 's/;/,/g' ${FILE_NAME}
done
