FROM golang:1.8
WORKDIR /opt/beacon
ARG ARTIFACT_URL
RUN wget $ARTIFACT_URL
RUN tar xvzf beacon-linux-amd64.tar.gz
ENTRYPOINT ["./bin/beacon-api"]
CMD ["--help"]
