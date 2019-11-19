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
    sudo faas-cli remove gci-proxy-resolver
    sudo faas-cli remove listfiller-gci
    sudo faas-cli remove listfiller-nogci
    sleep 1
    if [ "$flag" = "gci" ]
    then
        cd ../gci-proxy-resolver/
        sudo faas-cli deploy -f gci-proxy-resolver.yml
        sleep 1
        cd ../functions/listfiller-gci-func/
        sudo faas-cli up -f listfiller-gci.yml
        sudo docker service update --publish-add published=31123,target=8080 listfiller-gci

    else
        cd ../functions/listfiller-nogci-func/
        sudo faas-cli up -f listfiller-nogci.yml

    fi
    cd ../../experiment
    sleep 3

    FILE_NAME="${RESULTS_PATH}${FUNC}-${flag}${EXPID}.csv"
    echo -e "id;body;status;latency" > ${FILE_NAME}
    LOG_PATH=${RESULTS_PATH}${FUNC}-${flag}${EXPID}/
    mkdir -p ${LOG_PATH}
    for i in `seq 1 ${REQS}`
    do
        curl_return=$(curl -X GET -s -w ';%{http_code};%{time_total}\n' ${TARGET}${FUNC}-${flag})
        check=$(echo "${i};${curl_return}" | grep -v ";200;" | wc -l)
        if [ $check -gt 0 ];
            then
                container=$(sudo docker ps -f name=listfiller-${flag} --format "{{.ID}}")
                for id in ${container}; 
                do 
                    docker cp "${id}:/home/app/gc.log" "${LOG_PATH}gc-${id}.log"
                    docker cp "${id}:/home/app/proxy-stdout.log" "${LOG_PATH}proxy-stdout-${id}.log"
                    docker cp "${id}:/home/app/proxy-stderr.log" "${LOG_PATH}proxy-stderr-${id}.log"
                    docker logs ${id} >${LOG_PATH}stdout-${id}.log 2>${LOG_PATH}stderr-${id}.log
                done
                echo "${i};${curl_return}" >> ${FILE_NAME}
                break
            else
                echo "${i};${curl_return}" >> ${FILE_NAME}
        fi
    done
    container=$(sudo docker ps -f name=listfiller-${flag} --format "{{.ID}}")
    for id in ${container}; 
    do 
        docker cp "${id}:/home/app/gc.log" "${LOG_PATH}gc-${id}.log"
        docker cp "${id}:/home/app/proxy-stdout.log" "${LOG_PATH}proxy-stdout-${id}.log"
        docker cp "${id}:/home/app/proxy-stderr.log" "${LOG_PATH}proxy-stderr-${id}.log"
        docker logs ${id} >${LOG_PATH}stdout-${id}.log 2>${LOG_PATH}stderr-${id}.log
    done
    sed -i 's/,/./g' ${FILE_NAME}
    sed -i 's/;/,/g' ${FILE_NAME}
done
for f in $(docker service ps -q $service);do docker inspect --format '{{.NodeID}} {{.Status.ContainerStatus.ContainerID}}' $f; done
