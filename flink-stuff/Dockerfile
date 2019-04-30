FROM flink:1.7.2-hadoop26-scala_2.11-alpine
MAINTAINER sashidhar guntury <sguntury@gmail.com>

ENV FLINK_HOME="/opt/flink"
ENV FLINK_LIB_DIR=${FLINK_HOME}/lib
ENV FLINK_OPT_DIR=${FLINK_HOME}/opt
ENV FLINK_BIN_DIR=${FLINK_HOME}/bin
ENV PATH=${FLINK_BIN_DIR}:${PATH}

## COPY flink-conf.yaml ${FLINK_HOME}/conf
## COPY core-site.xml ${FLINK_HOME}/conf
## COPY log4j-console.properties ${FLINK_HOME}/conf

COPY entrypoint.sh /usr/src/app/

# RUN groupadd -r hadoop && useradd --no-log-init -r -g hadoop yarn
# USER yarn

ENV SBT_VERSION=1.2.4
## install sbt
RUN apk add --no-cache curl && \
    curl -fsL https://github.com/sbt/sbt/releases/download/v$SBT_VERSION/sbt-$SBT_VERSION.tgz | tar xfz - -C /usr/local && \
    ln -s /usr/local/sbt/bin/* /usr/local/bin/ && \
    apk del curl

EXPOSE 8081 6123

ENTRYPOINT ["/usr/src/app/entrypoint.sh"]

CMD []
