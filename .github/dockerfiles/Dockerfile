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
	libncurses5-dev m4 zlib1g-dev bc libelf-dev

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

# Build contest server
WORKDIR /root/contest
COPY . .
RUN --mount=type=ssh go build ./cmds/contest

#fwhunt dependencies 
RUN --mount=type=ssh git clone https://github.com/binarly-io/fwhunt-scan.git /root/fwhunt-scan
RUN --mount=type=ssh git clone https://github.com/binarly-io/FwHunt.git /root/fwhunt-scan/FwHunt-rules

FROM ubuntu:22.04

COPY --from=builder /root/contest/contest /usr/local/bin/contest
COPY --from=builder /root/fwhunt-scan /fwhunt-scan

RUN apt-get update
RUN apt-get install -y qemu-system pip curl

# install fwhunt
WORKDIR /fwhunt-scan
RUN echo 'deb http://download.opensuse.org/repositories/home:/RizinOrg/xUbuntu_22.04/ /' | tee /etc/apt/sources.list.d/home:RizinOrg.list
RUN curl -fsSL https://download.opensuse.org/repositories/home:RizinOrg/xUbuntu_22.04/Release.key | gpg --dearmor | tee /etc/apt/trusted.gpg.d/home_RizinOrg.gpg > /dev/null
RUN apt-get update --allow-insecure-repositories
RUN apt install rizin -y --allow-unauthenticated
RUN pip install -r requirements.txt

# Add SSH Keys for the DUTs
ADD .github/assets/ssh-keys/fwci /root/.ssh/fwci
ADD .github/assets/ssh-keys/fwci.pub /root/.ssh/fwci.pub

# Add ws-proxy-client for HWaaS
ADD .github/assets/ws-proxy-client /root/hwaas/ws-proxy-client

# Start everything
WORKDIR /
CMD contest -dbURI=""