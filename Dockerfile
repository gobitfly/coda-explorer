FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git mercurial gcc npm
ADD . /src
RUN cd /src && make -B all

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/bin /app/
CMD []