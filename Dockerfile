FROM heroku/cedar:14
MAINTAINER Heroku Build & Packaging Team <build-and-packaging@heroku.com>
COPY . /app
WORKDIR /app
ENV HOME /app
ENV PATH $PATH:$HOME/bin
RUN mkdir -p /var/lib/buildpack /var/cache/buildpack
RUN git clone --depth 1 https://github.com/heroku/heroku-buildpack-go.git /var/lib/buildpack
RUN /var/lib/buildpack/bin/compile /app /var/cache/buildpack
