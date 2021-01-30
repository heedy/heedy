FROM node AS frontend

# We don't need to clean up after this because this build stage is later discarded
RUN apt-get update && apt-get install -y rsync

WORKDIR /frontend
COPY . .

RUN make frontend


# We use the alpine tag to be able to execute on the alpine image later
FROM golang AS backend

WORKDIR /backend
COPY --from=frontend /frontend .

RUN make heedy && chmod +x ./heedy


FROM python:3.9-slim-buster

WORKDIR /data
WORKDIR /heedy

ENV HOME=/data
COPY --from=backend /backend/heedy .

RUN chgrp -R 0 /heedy /data && chmod -R g=u /heedy /data
USER 12938

EXPOSE 1324

CMD [ "./heedy" ]
