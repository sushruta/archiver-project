FROM asappinc/ubuntu_go111

ENV GOPATH=/gobits
RUN mkdir -p /gobits

COPY ./ /gobits

RUN cd /gobits/src/github.com/asappinc/challenge-di-eng && dep ensure

WORKDIR /gobits/src
RUN go install github.com/asappinc/challenge-di-eng/cmd/archiver
RUN go install github.com/asappinc/challenge-di-eng/cmd/echoconsume
RUN go install github.com/asappinc/challenge-di-eng/cmd/produce
