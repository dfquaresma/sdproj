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
    echo -e "id;status;latency" > ${FILE_NAME}
    for i in `seq 1 ${REQS}`
    do
        curl_return=$(curl -X GET -o /dev/null -s -w '${i};%{http_code};%{time_total}\n' ${TARGET}${FUNC}-${flag})
        check=$(echo ${curl_return} | grep -v ";200;" | wc -l)
        if [ $check -gt 0 ]; 
            then
                TAG=${RESULTS_PATH}${FUNC}-${flag}${EXPID}
                container=$(sudo docker ps -f name=listfiller-${flag} --format "{{.ID}}")
                for id in ${container}; 
                    do 
                        docker cp "${id}:/home/app/gc.log" "${TAG}-gc-${id}.log"
                        docker cp "${id}:/home/app/proxy-stdout.log" "${TAG}-proxy-stdout-${id}.log"
                        docker cp "${id}:/home/app/proxy-stderr.log" "${TAG}-proxy-stderr-${id}.log"
                        docker logs ${id} >${TAG}-stdout-${id}.log 2>${TAG}-stderr-${id}.log
                    done
                exit
            else
                echo ${curl_return} >> ${FILE_NAME}
        fi
    done
    sed -i 's/,/./g' ${FILE_NAME}
    sed -i 's/;/,/g' ${FILE_NAME}
done
for f in $(docker service ps -q $service);do docker inspect --format '{{.NodeID}} {{.Status.ContainerStatus.ContainerID}}' $f; done