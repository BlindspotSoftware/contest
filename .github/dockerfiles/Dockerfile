# syntax=docker/dockerfile:1.0.0-experimental
FROM ubuntu:22.04 AS builder

ARG TARGETARCH=amd64

# Install dependencies
RUN apt-get update
RUN apt-get -y upgrade

ENV DEBIAN_FRONTEND=noninteractive

# Install packages
RUN apt-get install -y git curl tzdata\ 
	bison build-essential curl flex git gnat\
	libncurses5-dev m4 zlib1g-dev bc libelf-dev qemu-system

RUN dpkg-reconfigure --frontend noninteractive tzdata

# Install Go
RUN curl -L https://golang.org/dl/go1.19.5.linux-${TARGETARCH}.tar.gz | tar xzf -
ENV PATH=$PATH:/go/bin

RUN ln -s /lib/x86_64-linux-gnu/libc.so.6 /lib/libc.so.6

# SSH Magic
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN touch /root/.ssh/config
RUN echo "StrictHostKeyChecking no" > /root/.ssh/config
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

# Clone contest server
RUN --mount=type=ssh git clone https://github.com/firmwareci/contest.git /root/contest

WORKDIR /root/contest
RUN --mount=type=ssh go build ./cmds/contest

FROM ubuntu:22.04

COPY --from=builder /root/contest/contest /usr/local/bin/contest

# Start everything
CMD contest -dbURI=""