FROM asappinc/ubuntu_go111

ENV GOPATH=/gobits
RUN mkdir -p /gobits

COPY ./ /gobits

RUN cd /gobits/src/github.com/asappinc/challenge-di-eng && dep ensure