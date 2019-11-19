FROM adoptopenjdk/openjdk11:jdk-11.0.2.9-slim as builder

RUN apt-get update \
    && apt-get install -y apt-utils curl sudo unzip \
    && useradd -m app && echo "app:app" | chpasswd && adduser app sudo \
    && mkdir -p /opt/ && cd /opt/ \
    && echo "Downloading gradle.." \
    && curl -SfL https://services.gradle.org/distributions/gradle-5.3-bin.zip -o gradle-bin.zip \
    && unzip gradle-bin.zip -d /opt/ \
    && rm gradle-bin.zip

# Export some environment variables
ENV GRADLE_HOME=/opt/gradle-5.3/
ENV PATH=$PATH:$GRADLE_HOME/bin

RUN mkdir -p /home/app/libs

ENV GRADLE_OPTS="-Dorg.gradle.daemon=false"
WORKDIR /home/app

COPY . /home/app/

RUN gradle build --stacktrace

FROM adoptopenjdk/openjdk11:jdk-11.0.2.9-slim as ship
RUN apt-get update \
    && apt-get install -y apt-utils curl unzip bash build-essential \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas-incubator/of-watchdog/releases/download/0.4.6/of-watchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

RUN useradd -m app && echo "app:app" | chpasswd && adduser app sudo
WORKDIR /home/app
COPY --from=builder /home/app/entrypoint/build/distributions/entrypoint-1.0.zip ./entrypoint-1.0.zip
RUN unzip ./entrypoint-1.0.zip

RUN chown -R app /home/app
USER app

ENV entrypoint_port="8082"
ENV upstream_url="http://127.0.0.1:8081"
ENV mode="http"
ENV CLASSPATH="/home/app/entrypoint-1.0/lib/*"

RUN mkdir "/home/app/gci"
COPY gci/* /home/app/gci/
ENV PATH_TO_GCI="/home/app/gci"
RUN bash ${PATH_TO_GCI}/libgcso-entrypoint.sh

ENV fprocess="java -cp /home/app/entrypoint-1.0/lib/* -server -Xms1024m -Xmx1024m -XX:+UnlockExperimentalVMOptions -XX:+UseSerialGC -XX:NewRatio=1 -Xlog:gc:file=/home/app/gc.log -Djvmtilib=/home/app/gci/libgc.so -javaagent:/home/app/gci/gciagent-0.1-jar-with-dependencies.jar=8500 com.openfaas.entrypoint.App"
EXPOSE 8080

HEALTHCHECK --interval=2s CMD [ -e /tmp/.lock ] || exit 1

CMD ["sh","-c","nohup ./gci/gci-proxy --port=8081 --use_mesh=true --target=127.0.0.1:8082 --service_name=listfiller-gci --mesh_resolver_url=http://150.165.75.122:8080/function/gci-proxy-resolve --gci_target=127.0.0.1:8500 --ygen=104857600 >/home/app/proxy-stdout.log 2>/home/app/proxy-stderr.log & fwatchdog"]
