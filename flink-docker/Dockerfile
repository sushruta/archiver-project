FROM openjdk:8-jdk-stretch
MAINTAINER sashidhar guntury <sguntury@gmail.com>

RUN apt-get update && \
    apt-get -y install openssh-server sudo gzip bash curl tar

RUN ssh-keygen -t rsa -P '' -f ~/.ssh/id_rsa && \
    cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys && \
    chmod 0600 ~/.ssh/authorized_keys

ADD conf/ssh_config /root/.ssh/config
RUN chmod 600 /root/.ssh/config && \
    chown root:root /root/.ssh/config

RUN sed -i s/#PermitRootLogin.*/PermitRootLogin\ yes/ /etc/ssh/sshd_config && \
    echo "root:root" | chpasswd && \
    sed -ie 's/#Port 22/Port 2122/g' /etc/ssh/sshd_config && \
    sed -ri 's/#HostKey \/etc\/ssh\/ssh_host_key/HostKey \/etc\/ssh\/ssh_host_key/g' /etc/ssh/sshd_config && \
    sed -ir 's/#HostKey \/etc\/ssh\/ssh_host_rsa_key/HostKey \/etc\/ssh\/ssh_host_rsa_key/g' /etc/ssh/sshd_config && \
    sed -ir 's/#HostKey \/etc\/ssh\/ssh_host_dsa_key/HostKey \/etc\/ssh\/ssh_host_dsa_key/g' /etc/ssh/sshd_config && \
    sed -ir 's/#HostKey \/etc\/ssh\/ssh_host_ecdsa_key/HostKey \/etc\/ssh\/ssh_host_ecdsa_key/g' /etc/ssh/sshd_config && \
    sed -ir 's/#HostKey \/etc\/ssh\/ssh_host_ed25519_key/HostKey \/etc\/ssh\/ssh_host_ed25519_key/g' /etc/ssh/sshd_config && \
    /usr/bin/ssh-keygen -A

RUN curl -s https://archive.apache.org/dist/hadoop/common/hadoop-2.8.3/hadoop-2.8.3.tar.gz | tar -xz -C /usr/local/ && \
    mv /usr/local/hadoop-2.8.3 /usr/local/hadoop

ENV HADOOP_HOME=/usr/local/hadoop
ENV FLINK_HOME=/opt/flink
ENV HADOOP_CONF_DIR=$HADOOP_HOME/etc/hadoop \
    FLINK_URL=http://apache.claz.org/flink/flink-1.8.0/flink-1.8.0-bin-scala_2.11.tgz \
    FLINK_HADOOP_URL=https://repo.maven.apache.org/maven2/org/apache/flink/flink-shaded-hadoop2-uber/2.8.3-1.8.0/flink-shaded-hadoop2-uber-2.8.3-1.8.0.jar \
    FLINK_CONF_DIR=$FLINK_HOME/conf \
    PATH=$HADOOP_HOME/bin:$HADOOP_HOME/sbin:$FLINK_HOME/bin:$PATH

RUN sed -ir 's/export JAVA_HOME=${JAVA_HOME}/export JAVA_HOME=\/docker-java-home/g' $HADOOP_CONF_DIR/hadoop-env.sh
ADD conf/core-site.xml $HADOOP_CONF_DIR/core-site.xml
ADD conf/hdfs-site.xml $HADOOP_CONF_DIR/hdfs-site.xml
ADD conf/mapred-site.xml $HADOOP_CONF_DIR/mapred-site.xml
ADD conf/yarn-site.xml $HADOOP_CONF_DIR/yarn-site.xml

RUN curl -s $FLINK_URL | tar -xz -C /opt/ && \
    mv /opt/flink-1.8.0 /opt/flink && \
    curl -s $FLINK_HADOOP_URL -o $FLINK_HOME/lib/hadoop.jar

COPY conf/flink-conf.yaml $FLINK_CONF_DIR/flink-conf.yaml
COPY conf/flink-core-site.xml $FLINK_CONF_DIR/core-site.yaml
COPY conf/log4j-console.properties $FLINK_CONF_DIR/log4j-console.properties
RUN echo "taskmanager1" >  $FLINK_CONF_DIR/slaves && \
    echo "taskmanager2" >> $FLINK_CONF_DIR/slaves

## ENV SBT_VERSION=1.2.4
## RUN apk add --no-cache curl && \
##     curl -fsL https://github.com/sbt/sbt/releases/download/v$SBT_VERSION/sbt-$SBT_VERSION.tgz | tar xfz - -C /usr/local && \
##     ln -s /usr/local/sbt/bin/* /usr/local/bin/ && \
##     apk del curl
## 
## ADD app /usr/src/app
## WORKDIR /usr/src/app
## RUN sbt clean compile assembly

ADD scripts/entrypoint.sh /bin/
RUN chmod +x /bin/entrypoint.sh

EXPOSE 8081 6121 6122 6123

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD []
