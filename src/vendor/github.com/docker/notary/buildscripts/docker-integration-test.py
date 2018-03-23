"""
"vendors" notary into docker and runs integration tests - then builds the
docker client binary with an API version compatible with the existing
daemon

Usage:
python docker-integration-test.py

This assumes that your docker directory is in $GOPATH/src/github.com/docker/docker
and your notary directory, irrespective of where this script is located, is
at $GOPATH/src/github.com/docker/notary.
"""
from __future__ import print_function
import os
import re
import shutil
import subprocess
import sys

def from_gopath(gopkg):
    """
    Gets the location of the go source given go package, based on the $GOPATH.
    """
    gopaths = os.getenv("GOPATH")
    for path in gopaths.split(":"):
        maybe_path = os.path.abspath(os.path.expanduser(os.path.join(
            path, "src", *gopkg.split("/"))))
        if os.path.isdir(maybe_path):
            return maybe_path
    return ""


DOCKER_DIR = from_gopath("github.com/docker/docker")
NOTARY_DIR = from_gopath("github.com/docker/notary")


def fake_vendor():
    """
    "vendors" notary into docker by copying all of notary into the docker
    vendor directory - also appending several lines into the Dockerfile because
    it pulls down notary from github and builds the binaries
    """
    docker_notary_relpath = "vendor/src/github.com/docker/notary"
    docker_notary_abspath = os.path.join(DOCKER_DIR, docker_notary_relpath)

    print("copying notary ({0}) into {1}".format(NOTARY_DIR, docker_notary_abspath))

    def ignore_dirs(walked_dir, _):
        """
        Don't vendor everything, particularly not the docker directory
        recursively, if it happened to be in the notary directory
        """
        if walked_dir == NOTARY_DIR:
            return [".git", ".cover", "docs", "bin"]
        elif walked_dir == os.path.join(NOTARY_DIR, "fixtures"):
            return ["compatibility"]
        return []

    if os.path.exists(docker_notary_abspath):
        shutil.rmtree(docker_notary_abspath)
    shutil.copytree(
        NOTARY_DIR, docker_notary_abspath, symlinks=True, ignore=ignore_dirs)

    # hack this because docker/docker's Dockerfile checks out a particular version of notary
    # based on a tag or SHA, and we want to build based on what was vendored in
    dockerfile_addition = ("\n"
        "RUN set -x && "
        "export GO15VENDOREXPERIMENT=1 && "
        "go build -o /usr/local/bin/notary-server github.com/docker/notary/cmd/notary-server &&"
        "go build -o /usr/local/bin/notary github.com/docker/notary/cmd/notary")

    with open(os.path.join(DOCKER_DIR, "Dockerfile")) as dockerfile:
        text = dockerfile.read()

    if not text.endswith(dockerfile_addition):
        with open(os.path.join(DOCKER_DIR, "Dockerfile"), 'a+') as dockerfile:
            dockerfile.write(dockerfile_addition)

    # hack the makefile so that we tag the built image as something else so we
    # don't interfere with any other docker test builds
    with open(os.path.join(DOCKER_DIR, "Makefile"), 'r') as makefile:
        makefiletext = makefile.read()

    with open(os.path.join(DOCKER_DIR, "Makefile"), 'wb') as makefile:
        image_name = os.getenv("DOCKER_TEST_IMAGE_NAME", "notary-docker-vendor-test")
        text = re.sub("^DOCKER_IMAGE := .+$", "DOCKER_IMAGE := {0}".format(image_name),
                      makefiletext, 1, flags=re.M)
        makefile.write(text)

def run_integration_test():
    """
    Presumes that the fake vendoring has already happened - this runs the
    integration tests.
    """
    env = os.environ.copy()
    env["TESTFLAGS"] = '-check.f DockerTrustSuite*'
    subprocess.check_call(
        "make test-integration-cli".split(), cwd=DOCKER_DIR, env=env)

if __name__ == "__main__":
    if len(sys.argv) > 1:
        print("\nWarning: Ignoring all extra arguments: {0}".format(" ".join(sys.argv[1:])))
        print("\nUsage: python {0}\n\n".format(sys.argv[0]))
    if DOCKER_DIR == "":
        print("ERROR: Could not find github.com/docker/docker in your GOPATH='{0}'"
              .format(os.getenv("GOPATH")))
        sys.exit(1)
    if NOTARY_DIR == "":
        print("ERROR: Could not find github.com/docker/notary in your GOPATH='{0}'"
              .format(os.getenv("GOPATH")))
        sys.exit(1)
    fake_vendor()
    run_integration_test()
