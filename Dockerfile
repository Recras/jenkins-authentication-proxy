FROM google/golang

WORKDIR /gopath/src/jenkins-authentication-proxy
ADD . /gopath/src/jenkins-authentication-proxy/

RUN go get jenkins-authentication-proxy

ENTRYPOINT ["/gopath/bin/jenkins-authentication-proxy"]
