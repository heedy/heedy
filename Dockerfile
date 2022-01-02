# syntax=docker/dockerfile:1

FROM node AS frontend

# We don't need to clean up after this because this build stage is later discarded
RUN apt-get update && apt-get install -y rsync

WORKDIR /frontend
COPY . .

RUN make frontend


FROM golang AS backend

WORKDIR /backend
COPY --from=frontend /frontend .

ARG VERSION

RUN if [[ -z "$VERSION" ]] ; then make heedy ; else make heedy VERSION=$VERSION ; fi

FROM python:3.10-slim-bullseye

# Things like Jupyter really want a home directory to write their own stuff to.
WORKDIR /home
ENV HOME=/home

# The folder which will hold the heedy database
WORKDIR /data

WORKDIR /

COPY --from=backend /backend/heedy .

# Grant docker user group access
RUN chgrp -R 0 /data /home && chmod -R g=u /data /home

USER 12938

EXPOSE 1324

CMD [ "/heedy","run", "/data", "--create" ]
