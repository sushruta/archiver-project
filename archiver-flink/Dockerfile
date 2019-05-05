FROM flink:1.8.0-scala_2.11
MAINTAINER sashidhar guntury <sguntury@gmail.com>

ENV PATH=$FLINK_HOME/bin:$PATH \
    SBT_VERSION=1.2.4 \
    ASSEMBLY_JAR_PATH=/usr/src/app/archiver.jar \
    JOB_CLASS_NAME=com.asappinc.Archiver

## install sbt
RUN curl -fsL https://github.com/sbt/sbt/releases/download/v$SBT_VERSION/sbt-$SBT_VERSION.tgz | tar xfz - -C /usr/local && \
    ln -s /usr/local/sbt/bin/* /usr/local/bin/

WORKDIR /usr/src/app
ADD project /usr/src/app/project
ADD build.sbt /usr/src/app
RUN sbt sbtVersion
ADD src /usr/src/app/src
RUN sbt clean compile assembly

#ADD scripts/entrypoint.sh /bin/entrypoint.sh
#RUN chmod +x /bin/entrypoint.sh
#
CMD ["flink", "run", "-m", "jobmanager:8081", "-c", "com.asappinc.Archiver", "archiver.jar"]
#ENTRYPOINT ["/bin/entrypoint.sh"]
#CMD []
