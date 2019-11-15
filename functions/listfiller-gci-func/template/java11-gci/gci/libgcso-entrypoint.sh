#!/bin/bash

date
set -x

gcc -shared -fpic -I"${JAVA_HOME}/include" -I"${JAVA_HOME}/include/linux" ${PATH_TO_GCI}/com_gcinterceptor_core_GC.c -o ${PATH_TO_GCI}/libgc.so || exit $?
