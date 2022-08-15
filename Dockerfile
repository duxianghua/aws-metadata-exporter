FROM alpine:3.16

WORKDIR /app

COPY ./bin/aws-metadata-exporter .

EXPOSE 8080

ENV PORT 9091
ENV HOST 0.0.0.0

ENTRYPOINT ["./aws-metadata-exporter"]
