#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

FROM golang:stretch

# Configure apt
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get -y install --no-install-recommends apt-utils 2>&1

# Install git, process tools
RUN apt-get -y install git procps wget nano curl

# Install C++ tools
RUN apt-get -y install build-essential cmake cppcheck valgrind pkg-config lsb-release 

# Clean up
RUN apt-get autoremove -y \
    && apt-get clean -y \ 
    && rm -rf /var/lib/apt/lists/*

# Set the default shell to bash instead of sh
ENV SHELL /bin/bash