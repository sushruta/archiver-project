FROM openjdk:8-jre-alpine
MAINTAINER sashidhar guntury <sguntury@gmail.com>

RUN apk add --no-cache gzip openssh bash curl tar sudo

RUN ssh-keygen -t rsa -P '' -f ~/.ssh/id_rsa && \
    cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys && \
    chmod 0600 ~/.ssh/authorized_keys

RUN curl -s http://apache.mirrors.hoobly.com/hadoop/common/hadoop-2.9.2/hadoop-2.9.2.tar.gz | tar -xz -C /usr/local/ && \
    mv /usr/local/hadoop-2.9.2 /usr/local/hadoop

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

ENV HADOOP_PREFIX=/usr/local/hadoop
ENV HADOOP_HOME=$HADOOP_PREFIX \
HADOOP_COMMON_HOME=$HADOOP_PREFIX \
HADOOP_HDFS_HOME=$HADOOP_PREFIX \
HADOOP_MAPRED_HOME=$HADOOP_PREFIX \
HADOOP_YARN_HOME=$HADOOP_PREFIX \
HADOOP_CONF_DIR=$HADOOP_PREFIX/etc/hadoop \
YARN_CONF_DIR=$HADOOP_PREFIX/etc/hadoop \
PATH=$HADOOP_PREFIX/bin:$HADOOP_PREFIX/sbin:$PATH

RUN sed -ir 's/export JAVA_HOME=${JAVA_HOME}/export JAVA_HOME=\/usr\/lib\/jvm\/java-1.8-openjdk\/jre/g' $HADOOP_CONF_DIR/hadoop-env.sh
ADD conf/core-site.xml $HADOOP_CONF_DIR/core-site.xml
ADD conf/hdfs-site.xml $HADOOP_CONF_DIR/hdfs-site.xml
## ADD conf/mapred-site.xml $HADOOP_CONF_DIR/mapred-site.xml
## ADD conf/yarn-site.xml $HADOOP_CONF_DIR/yarn-site.xml

ADD scripts/entrypoint.sh /bin/entrypoint.sh
RUN chmod +x /bin/entrypoint.sh
# $HADOOP_CONF_DIR/hadoop-env.sh

# hdfs ports
EXPOSE \
50010 50020 50070 50075 50090 8020 9000 \
# mapred ports
10020 19888 \
# yarn ports
8030 8031 8032 8033 8040 8042 8088 \
#other misc ports
49707 2122 54311

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD []
