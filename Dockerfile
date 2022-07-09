FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
ENV GOPROXY https://goproxy.cn,direct
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/distribute_tf_worker .

FROM alpine
COPY --from=build /out/distribute_tf_worker /bin
CMD ["distribute_tf_worker"]
