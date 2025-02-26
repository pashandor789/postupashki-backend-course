FROM golang:1.22.3-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add --no-cache make

RUN go build main.go

FROM alpine AS runner

WORKDIR /home/danil/project_postup/postupashki-backend-course/lec2example

RUN apk add --no-cache curl

COPY --from=build /build/main ./main

CMD ["/home/danil/project_postup/postupashki-backend-course/lec2example/main" ]
