FROM golang:1.15 AS build
WORKDIR /app
RUN git clone https://github.com/polynetwork/polygon-relayer.git  && \
    cd polygon-relayer && \
    go build -o run_polygon_relayer main.go

FROM ubuntu:18.04
WORKDIR /app
COPY ./config.json config.json
COPY --from=build /app/polygon-relayer/run_polygon_relayer run_polygon_relayer
CMD ["/bin/bash"]