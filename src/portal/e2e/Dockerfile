# FROM docker:27-dind
#
# # Install Node & Playwright dependencies
# RUN apk add --no-cache \
#     nodejs \
#     npm \
#     bash \
#     curl \
#     libc6-compat \
#     chromium \
#     chromium-chromedriver
#
# # Install Playwright + browsers
# RUN npm install -g playwright@1.56.1 && \
#     playwright install --with-deps
#
# # Create work directory
# WORKDIR /tests
#
# # Copy your tests
# COPY package.json package-lock.json ./
# RUN npm install
#
# COPY . ./tests
#
# # Docker daemon entrypoint + playwright test runner
# CMD ["npx", "playwright", "test", "--reporter=html"]
# FROM node:20-bookworm
#
# RUN npx -y playwright@1.56.1 install --with-deps
#
# WORKDIR /app
#
# COPY e2e ./
#
# CMD ["npx", "playwright", "test", "--reporter=html"]

# FROM mcr.microsoft.com/playwright:v1.56.1-noble
FROM node:20-bookworm

# RUN npx -y playwright@1.56.1 install --with-deps

# 1. Install Docker CLI:
# We install the Docker Command Line Interface (CLI) to allow the container
# to execute docker commands (e.g., to spin up your application under test).
# Install necessary packages for Docker
RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    sudo

# Add Docker's GPG key
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Add Docker repository
RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
RUN apt-get update && apt-get install -y \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-buildx-plugin \
    docker-compose-plugin

# Create a non-root user (optional, but good practice)
# RUN useradd -ms /bin/bash dockeruser
# RUN usermod -aG docker dockeruser

# Switch to the non-root user
# USER dockeruser
# RUN apt-get update && \
#     apt-get install -y --no-install-recommends \
#     ca-certificates \
#     curl \
#     gnupg \
#     lsb-release && \
#     # Add Docker's official GPG key and repository
#     mkdir -m 0755 -p /etc/apt/keyrings && \
#     curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg && \
#     echo \
#     "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
#     $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
#     # Install the Docker CLI client
#     apt-get update && \
#     apt-get install -y docker-ce-cli && \
#     # Cleanup to keep the image small
#     rm -rf /var/lib/apt/lists/*

# Install Playwright dependencies (if not using the official Playwright base image)
# If you stick with the Playwright base image above, this step is often unnecessary.
# If you are using package.json dependencies:

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install @playwright/test
RUN npx playwright install --with-deps

COPY . .

# Command to run the tests
CMD ["npx", "playwright", "test", "--reporter=html"]
