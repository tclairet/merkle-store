FROM public.ecr.aws/docker/library/golang:1.21.0-alpine3.18

WORKDIR /merkle-store
COPY . .
RUN go mod download

RUN cd cmd/server && go build -o server .

EXPOSE 3333

CMD ["./cmd/server/server"]