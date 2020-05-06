#!/bin/bash

set +e
set -o noglob

if [ "$1" == "" ];then
echo "This shell will push specific image to registry server."
echo "Usage: #./pushimage [image tag] [registry username] [registry password]  [registry server]"
exit 1
fi

#
# Set Colors
#

bold=$(tput bold)
underline=$(tput sgr 0 1)
reset=$(tput sgr0)

red=$(tput setaf 1)
green=$(tput setaf 76)
white=$(tput setaf 7)
tan=$(tput setaf 202)
blue=$(tput setaf 25)

#
# Headers and Logging
#

underline() { printf "${underline}${bold}%s${reset}\n" "$@"
}
h1() { printf "\n${underline}${bold}${blue}%s${reset}\n" "$@"
}
h2() { printf "\n${underline}${bold}${white}%s${reset}\n" "$@"
}
debug() { printf "${white}%s${reset}\n" "$@"
}
info() { printf "${white}➜ %s${reset}\n" "$@"
}
success() { printf "${green}✔ %s${reset}\n" "$@"
}
error() { printf "${red}✖ %s${reset}\n" "$@"
}
warn() { printf "${tan}➜ %s${reset}\n" "$@"
}
bold() { printf "${bold}%s${reset}\n" "$@"
}
note() { printf "\n${underline}${bold}${blue}Note:${reset} ${blue}%s${reset}\n" "$@"
}


type_exists() {
  if [ $(type -P $1) ]; then
    return 0
  fi
  return 1
}

# Check variables
if [ -z $1 ]; then
  error "Please set the 'image' variable"
  exit 1
fi

if [ -z $2 ]; then
  error "Please set the 'username' variable"
  exit 1
fi

if [ -z $3 ]; then
  error "Please set the 'password' variable"
  exit 1
fi

if [ -z $4 ]; then
  info "Using default registry server (dockerhub)."
fi


# Check Docker is installed
if ! type_exists 'docker'; then
  error "Docker is not installed."
  info "Please install docker package."
  exit 1
fi

# Variables
IMAGE="$1"
USERNAME="$2"
PASSWORD="$3"
REGISTRY="$4"

set -e

# ----- Pushing image(s) -----
# see documentation :
#  - https://docs.docker.com/reference/commandline/cli/#login
#  - https://docs.docker.com/reference/commandline/cli/#push
#  - https://docs.docker.com/reference/commandline/cli/#logout
# ---------------------------

# Login to the registry
h2 "Login to the Docker registry"

DOCKER_LOGIN="docker login --username $USERNAME --password $PASSWORD $REGISTRY"
info "docker login --username $USERNAME --password *******"
DOCKER_LOGIN_OUTPUT=$($DOCKER_LOGIN)

if [ $? -ne 0 ]; then
  warn "$DOCKER_LOGIN_OUTPUT"
  error "Login to Docker registry $REGISTRY failed"
  exit 1
else
  success "Login to Docker registry $REGISTRY succeeded";
fi

# Push the docker image
h2 "Pushing image to Docker registry"

DOCKER_PUSH="docker push $IMAGE"
info "$DOCKER_PUSH"
DOCKER_PUSH_OUTPUT=$($DOCKER_PUSH)

if [ $? -ne 0 ];then
  warn $DOCKER_PUSH_OUTPUT
  error "Pushing image $IMAGE failed";
else
  success "Pushing image $IMAGE succeeded";
fi

# Logout from the registry
h2 "Logout from the docker registry"
DOCKER_LOGOUT="docker logout $REGISTRY"
DOCKER_LOGOUT_OUTPUT=$($DOCKER_LOGOUT)

if [ $? -ne 0 ]; then
  warn "$DOCKER_LOGOUT_OUTPUT"
  error "Logout from Docker registry $REGISTRY failed"
  exit 1
else
  success "Logout from Docker registry $REGISTRY succeeded"
fi
