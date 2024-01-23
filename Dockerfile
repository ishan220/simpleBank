#Build Stage
##base golangimage:tag
FROM golang:1.21-alpine3.18 AS Builder

###all files copied to WORKDIR
WORKDIR /app

####copyinh all file from current directory to working directory first is current and 
###second is work dir after previous command dir is changed to work directory
COPY . .

### -o <executable> <entrypoint file>
RUN go build -o main main.go

####Run Stage
FROM alpine:3.18
WORKDIR /app

### . represents the work dire set from above command ,and app/main is the path from the builder stage
COPY --from=Builder /app/main .
COPY app.env .
##this id for readme to inform about the port exposed by the service
EXPOSE 8080

###command at last will run to execute the executable
CMD ["/app/main"]