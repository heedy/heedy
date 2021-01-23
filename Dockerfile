FROM node AS frontend

# We don't need to clean up after this because this build stage is later discarded
RUN apt-get update && apt-get install -y rsync

WORKDIR /frontend
COPY . .

RUN make frontend


# We use the alpine tag to be able to execute on the alpine image later
FROM golang:alpine AS backend

WORKDIR /backend
COPY --from=frontend /frontend .

RUN apk add --update build-base && make heedy && chmod +x ./heedy


FROM alpine

WORKDIR /heedy
COPY --from=backend /backend/heedy .

EXPOSE 1324

CMD [ "./heedy" ]
