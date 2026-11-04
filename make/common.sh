#!/bin/bash
#docker version: 20.10.10+
#docker-compose version: 1.18.0+
#podman and podman-compose are also supported by the installer
#golang version: 1.12.0+

set +e
set -o noglob

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

set -e

function check_golang {
	if ! go version &> /dev/null
	then
		warn "No golang package in your environment. You should use golang docker image build binary."
		return
	fi

	# golang has been installed and check its version
	if [[ $(go version) =~ (([0-9]+)\.([0-9]+)([\.0-9]*)) ]]
	then
		golang_version=${BASH_REMATCH[1]}
		golang_version_part1=${BASH_REMATCH[2]}
		golang_version_part2=${BASH_REMATCH[3]}

		# the version of golang does not meet the requirement
		if [ "$golang_version_part1" -lt 1 ] || ([ "$golang_version_part1" -eq 1 ] && [ "$golang_version_part2" -lt 12 ])
		then
			warn "Better to upgrade golang package to 1.12.0+ or use golang docker image build binary."
			return
		else
			note "golang version: $golang_version"
		fi
	else
		warn "Failed to parse golang version."
		return
	fi
}

function check_docker {
	if ! docker --version &> /dev/null
	then
		error "Need to install docker(20.10.10+) first and run this script again."
		exit 1
	fi

	# docker has been installed and check its version
	if [[ $(docker --version) =~ (([0-9]+)\.([0-9]+)([\.0-9]*)) ]]
	then
		docker_version=${BASH_REMATCH[1]}
		docker_version_part1=${BASH_REMATCH[2]}
		docker_version_part2=${BASH_REMATCH[3]}

		note "docker version: $docker_version"
		# the version of docker does not meet the requirement
		if [ "$docker_version_part1" -lt 17 ] || ([ "$docker_version_part1" -eq 17 ] && [ "$docker_version_part2" -lt 6 ])
		then
			error "Need to upgrade docker package to 20.10.10+."
			exit 1
		fi
	else
		error "Failed to parse docker version."
		exit 1
	fi
}

function check_container_runtime {
	if [ -z "${CONTAINER_RUNTIME:-}" ]
	then
		if command -v docker &> /dev/null
		then
			CONTAINER_RUNTIME=docker
		elif command -v podman &> /dev/null
		then
			CONTAINER_RUNTIME=podman
		else
			error "Need to install Docker (20.10.10+) or Podman first and run this script again."
			exit 1
		fi
	fi

	case "$CONTAINER_RUNTIME" in
		docker)
			check_docker
			;;
		podman)
			if ! podman --version &> /dev/null
			then
				error "CONTAINER_RUNTIME is set to podman, but Podman is not installed."
				exit 1
			fi
			note "$(podman --version)"
			;;
		*)
			error "Unsupported container runtime: $CONTAINER_RUNTIME. Use docker or podman."
			exit 1
			;;
	esac
}

function check_dockercompose {
	if [ -n "${DOCKER_COMPOSE:-}" ]
	then
		if $DOCKER_COMPOSE version &> /dev/null
		then
			note "$($DOCKER_COMPOSE version)"
		elif $DOCKER_COMPOSE --version &> /dev/null
		then
			note "$($DOCKER_COMPOSE --version)"
		else
			error "The configured compose command is not available: $DOCKER_COMPOSE"
			exit 1
		fi
		return
	fi

	if [ "${CONTAINER_RUNTIME:-docker}" = "podman" ]
	then
		if podman-compose --version &> /dev/null
		then
			note "$(podman-compose --version)"
			DOCKER_COMPOSE="podman-compose"
		elif podman compose version &> /dev/null
		then
			note "$(podman compose version)"
			DOCKER_COMPOSE="podman compose"
		else
			error "Need to install podman-compose and run this script again."
			exit 1
		fi
		return
	fi

	# either docker compose plugin has been installed
	if docker compose version &> /dev/null
	then
		note "$(docker compose version)"
		DOCKER_COMPOSE="docker compose"

	# or docker-compose has been installed, check its version
	elif [[ $(docker-compose --version) =~ (([0-9]+)\.([0-9]+)([\.0-9]*)) ]]
	then
		docker_compose_version=${BASH_REMATCH[1]}
		docker_compose_version_part1=${BASH_REMATCH[2]}
		docker_compose_version_part2=${BASH_REMATCH[3]}

		note "docker-compose version: $docker_compose_version"
		# the version of docker-compose does not meet the requirement
		if [ "$docker_compose_version_part1" -lt 1 ] || ([ "$docker_compose_version_part1" -eq 1 ] && [ "$docker_compose_version_part2" -lt 18 ])
		then
			error "Need to upgrade docker-compose package to 1.18.0+."
			exit 1
		fi
	else
		error "Need to install docker-compose(1.18.0+) or a docker-compose plugin and run this script again."
		exit 1
	fi
}
