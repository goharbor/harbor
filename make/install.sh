#!/bin/bash

#docker version: 1.11.2 
#docker-compose version: 1.7.1 
#Harbor version: 0.4.0 

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
set +o noglob

usage=$'Please set hostname and other necessary attributes in harbor.cfg first. DO NOT use localhost or 127.0.0.1 for hostname, because Harbor needs to be accessed by external clients.
Please set --with-notary if needs enable Notary in Harbor, and set ui_url_protocol/ssl_cert/ssl_cert_key in harbor.cfg bacause notary must run under https.'
item=0

# notary is not enabled by default
with_notary=$false

while [ $# -gt 0 ]; do
        case $1 in
            --help)
            note "$usage"
            exit 0;;
            --with-notary)
            with_notary=true;;	 	
            *)
            note "$usage"
            exit 1;;
        esac
        shift || true
done

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $workdir

# The hostname in harbor.cfg has not been modified
if grep 'hostname = reg.mydomain.com' &> /dev/null harbor.cfg
then
	warn "$usage"
	exit 1
fi

function check_docker {
	if ! docker --version &> /dev/null
	then
		error "Need to install docker(1.10.0+) first and run this script again."
		exit 1
	fi
	
	# docker has been installed and check its version
	if [[ $(docker --version) =~ (([0-9]+).([0-9]+).([0-9]+)) ]]
	then
		docker_version=${BASH_REMATCH[1]}
		docker_version_part1=${BASH_REMATCH[2]}
		docker_version_part2=${BASH_REMATCH[3]}
		
		# the version of docker does not meet the requirement
		if [ "$docker_version_part1" -lt 1 ] || ([ "$docker_version_part1" -eq 1 ] && [ "$docker_version_part2" -lt 10 ])
		then
			error "Need to upgrade docker package to 1.10.0+."
			exit 1
		else
			note "docker version: $docker_version"
		fi
	else
		error "Failed to parse docker version."
		exit 1
	fi
}

function check_dockercompose {
	if ! docker-compose --version &> /dev/null
	then
		error "Need to install docker-compose(1.7.1+) by yourself first and run this script again."
		exit 1
	fi
	
	# docker-compose has been installed, check its version
	if [[ $(docker-compose --version) =~ (([0-9]+).([0-9]+).([0-9]+)) ]]
	then
		docker_compose_version=${BASH_REMATCH[1]}
		docker_compose_version_part1=${BASH_REMATCH[2]}
		docker_compose_version_part2=${BASH_REMATCH[3]}
		
		# the version of docker-compose does not meet the requirement
		if [ "$docker_compose_version_part1" -lt 1 ] || ([ "$docker_compose_version_part1" -eq 1 ] && [ "$docker_compose_version_part2" -lt 6 ])
		then
			error "Need to upgrade docker-compose package to 1.7.1+."
                        exit 1
		else
			note "docker-compose version: $docker_compose_version"
		fi
	else
		error "Failed to parse docker-compose version."
		exit 1
	fi
}

h2 "[Step $item]: checking installation environment ..."; let item+=1
check_docker
check_dockercompose

if [ -f harbor*.tar.gz ]
then
	h2 "[Step $item]: loading Harbor images ..."; let item+=1
	docker load -i ./harbor*.tar.gz
fi
echo ""

h2 "[Step $item]: preparing environment ...";  let item+=1
if [ -n "$host" ]
then
	sed "s/^hostname = .*/hostname = $host/g" -i ./harbor.cfg
fi
if [ $with_notary ]
then
	./prepare --with-notary
else
	./prepare
fi
echo ""

h2 "[Step $item]: checking existing instance of Harbor ..."; let item+=1
if [ $with_notary ]
then 
	if [ -n "$(docker-compose -f docker-compose.yml -f docker-compose.notary.yml ps -q)"  ]
	then
		note "stopping existing Harbor instance ..." 
		docker-compose -f docker-compose.yml -f docker-compose.notary.yml down -v
	fi
else
	if [ -n "$(docker-compose -f docker-compose.yml ps -q)" ]
	then
		note "stopping existing Harbor instance ..." 
		docker-compose -f docker-compose.yml down -v
	fi	
fi
echo ""

h2 "[Step $item]: starting Harbor ..."
if [ $with_notary ]
then
	docker-compose -f docker-compose.yml -f docker-compose.notary.yml up -d
else
	docker-compose -f docker-compose.yml up -d
fi

protocol=http
hostname=reg.mydomain.com

if [[ $(cat ./harbor.cfg) =~ ui_url_protocol[[:blank:]]*=[[:blank:]]*(https?) ]]
then
protocol=${BASH_REMATCH[1]}
fi

if [[ $(grep 'hostname[[:blank:]]*=' ./harbor.cfg) =~ hostname[[:blank:]]*=[[:blank:]]*(.*) ]]
then
hostname=${BASH_REMATCH[1]}
fi
echo ""

success $"----Harbor has been installed and started successfully.----

Now you should be able to visit the admin portal at ${protocol}://${hostname}. 
For more details, please visit https://github.com/vmware/harbor .
"
