FROM scalify/glide:0.13.2 as builder
WORKDIR /go/src/github.com/Scalify/website-content-watcher

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/website-content-watcher .


FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/Scalify/website-content-watcher/bin/website-content-watcher .
RUN chmod +x website-content-watcher
ENTRYPOINT ["./website-content-watcher"]
CMD ["watch"]
