# syntax=docker/dockerfile:1
FROM golang:1.19-alpine
ENV SERVER_PORT=8000
ENV MONGO_DB_USER=brguru90
ENV MONGO_DB_PASSWORD=guruven1
ENV MONGO_DB_HOST=cluster0.w8lcr.mongodb.net
ENV MONGO_DATABASE=cca
ENV MONGO_DB_PORT=27017
ENV REDIS_ADDR=redis://red-ch5obkcs3fvuob9msqqg:6379
ENV JWT_SECRET_KEY=default_dev_jwt_key_1234
ENV JWT_TOKEN_EXPIRE_IN_MINS=60
ENV ENABLE_REDIS_CACHE=true
ENV RESPONSE_CACHE_TTL_IN_SECS=300
ENV APP_ENV=development
ENV NODE_ENV=development
ENV GIN_MODE=debug
ENV DISABLE_COLOR=false
ENV FIREBASE_PUBLIC_API_KEY=AIzaSyAJwqFWjDt45O1kJVqRNt2SYt7aZZZsTFI
ENV FIREBASE_AUTH_DOMAIN=crud-eb4cc.firebaseapp.com
ENV FIREBASE_PROJECT_ID=crud-eb4cc
ENV PROTECTED_UPLOAD_PATH=uploads/private/
ENV PROTECTED_UPLOAD_PATH_ROUTE=private
ENV UNPROTECTED_UPLOAD_PATH=uploads/public/
ENV UNPROTECTED_UPLOAD_PATH_ROUTE=cdn
ENV RAZORPAY_KEY_ID=rzp_test_H8OTCJN1OWNZcp
ENV RAZORPAY_KEY_SECRET=EBYdW8qGNmCDrInXgTf4yfOm
ENV CGO_ENABLED=0
ENV GOOS=linux

ENV BUILDKIT_PROGRESS=plain
ENV SERVER_PATH=/web_app/


RUN echo $SERVER_PATH
RUN echo $SERVER_PORT

RUN mkdir -p $SERVER_PATH

WORKDIR $SERVER_PATH

COPY ./src ./src


RUN ls -lh ./src

RUN pwd
RUN ls  ./
RUN echo "---------"
RUN ls $SERVER_PATH/src


RUN go mod init cca
RUN go get ./src
RUN go get -u github.com/razorpay/razorpay-go
RUN go install github.com/razorpay/razorpay-go
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN go install github.com/swaggo/swag/cmd/swag
RUN ~/go/bin/swag init --dir src
RUN go build -v -o go_server src/main.go

EXPOSE $SERVER_PORT

ENTRYPOINT ["./go_server","-micro_service","cron_job"]