FROM golang:1.9
WORKDIR /opt/beacon
ARG ARTIFACT_URL
RUN wget -q $ARTIFACT_URL
RUN tar xvzf beacon-linux-amd64.tar.gz
RUN touch .env
RUN mkdir .keys && openssl genrsa -out ./.keys/private.pem 2048
EXPOSE 8080
ENTRYPOINT ["./bin/beacon-api"]
CMD ["--help"]
