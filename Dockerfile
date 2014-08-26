FROM google/golang

EXPOSE 8080

ENTRYPOINT ["/gopath/bin/jenkins-authentication-proxy"]

WORKDIR /gopath/src/jenkins-authentication-proxy
ADD . /gopath/src/jenkins-authentication-proxy/

RUN go get jenkins-authentication-proxy
