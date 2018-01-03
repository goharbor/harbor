#!/usr/bin/env python

"""
Run basic notary client tests against a server
"""

from __future__ import print_function

import argparse
from getpass import getpass
import inspect
import json
import os
from shutil import rmtree
from subprocess import CalledProcessError, PIPE, Popen, call
from tempfile import mkdtemp, mkstemp
from textwrap import dedent
from time import sleep, time
from uuid import uuid4

def reporoot():
    """
    Get the root of the git repo
    """
    return os.path.dirname(
        os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe()))))

# Returns the reponame and server name
def parse_args(args=None):
    """
    Parses the command line args for this command
    """
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=dedent("""
            Tests the notary client against a host.

            To test against a testing host without auth, just run without any arguments
            except maybe for the server; a random repo name will be generated.


            When running against Docker Hub, suggest usage like:

                python buildscripts/testclient.py \\
                    -s https://notary.docker.io \\
                    -r docker.io/username/reponame \\
                    -u username

            Note that especially for Docker Hub, the repo has to have been already
            created, else auth won't succeed.
            """))
    parser.add_argument(
        '-r', '--reponame', dest="reponame", type=str,
        help="The name of the repo - will be randomly generated if not provided")
    parser.add_argument(
        '-s', '--server', dest="server", type=str,
        help="Notary Server to connect to - defaults to https://notary-server:4443")
    parser.add_argument(
        '-u', '--username', dest="username", type=str,
        help="Username to use to log into the Notary Server (you will be asked for the password")
    parsed = parser.parse_args(args)

    return parsed.reponame, parsed.server, parsed.username

def cleanup(*paths):
    """
    Best effort removal the temporary paths, whether file or directory
    """
    for path in paths:
        try:
            os.remove(path)
        except OSError:
            pass
        else:
            continue

        try:
            rmtree(path)
        except OSError:
            pass

class Client(object):
    """
    Object that will run the notary client with the proper command lines
    """
    def __init__(self, notary_server, username_passwd=()):
        self.notary_server = notary_server
        self.username_passwd = username_passwd

        binary = os.path.join(reporoot(), "bin", "notary")
        self.env = os.environ.copy()
        self.env.update({
            "NOTARY_ROOT_PASSPHRASE": "root_ponies",
            "NOTARY_TARGETS_PASSPHRASE": "targets_ponies",
            "NOTARY_SNAPSHOT_PASSPHRASE": "snapshot_ponies",
            "NOTARY_DELEGATION_PASSPHRASE": "user_ponies",
        })

        if notary_server is None:
            self.client = [binary, "-D", "-c", "cmd/notary/config.json"]
        else:
            self.client = [binary, "-s", notary_server]

    def run(self, args, trust_dir, stdinput=None, username_passwd=None):
        """
        Runs the notary client in a subprocess, and returns the output
        """
        command = self.client + ["-d", trust_dir] + list(args)
        print("$ " + " ".join(command))

        # username password require newlines - EOF doesn't seem to do it.
        communicate_input = (tuple((x + "\n" for x in self.username_passwd))
                             if username_passwd is None else username_passwd)

        # Input comes before the username/password, and if there is a username
        # and password, we need a newline after the input.  Otherwise, just use
        # EOF (for instance if we're piping text to verify)
        if stdinput is not None:
            if communicate_input:
                communicate_input = (stdinput + "\n",) + communicate_input
            else:
                communicate_input = (stdinput,)

        _, filename = mkstemp()
        with open(filename, 'wb') as tempfile:
            process = Popen(command, env=self.env, stdout=tempfile, stdin=PIPE,
                            universal_newlines=True)
            for inp in communicate_input:
                process.stdin.write(inp)
            process.stdin.close()
            process.wait()

        with open(filename) as tempfile:
            output = tempfile.read()

        retcode = process.poll()
        cleanup(filename)
        print(output)
        if retcode:
            raise CalledProcessError(retcode, command, output=output)
        return output

class Tester(object):
    """
    Thing that runs the test
    """
    def __init__(self, repo_name, client):
        self.repo_name = repo_name
        self.client = client
        self.dir = mkdtemp(suffix="_main")

    def basic_repo_test(self, tempfile, tempdir):
        """
        Initialize a repo, add a target, ensure the target is readable
        """
        print("---- Initializing a repo, adding a target, and pushing ----\n")
        self.client.run(["init", self.repo_name], self.dir)
        self.client.run(["add", self.repo_name, "basic_repo_test", tempfile], self.dir)
        self.client.run(["publish", self.repo_name], self.dir)

        print("---- Listing and validating basic repo test targets ----\n")
        targets1 = self.client.run(["list", self.repo_name], self.dir)
        targets2 = self.client.run(["list", self.repo_name], tempdir)

        assert targets1 == targets2, "targets lists not equal: \n{0}\n{1}".format(
            targets1, targets2)
        assert "basic_repo_test" in targets1, "missing expected basic_repo_test: {0}".format(
            targets1)

        self.client.run(
            ["verify", self.repo_name, "basic_repo_test", "-i", tempfile, "-q"], self.dir,
            # skip username/password since this is an offline operation
            username_passwd=())

    def add_delegation_test(self, tempfile, tempdir):
        """
        Add a delegation to the repo - assumes the repo has already been initialized
        """
        print("---- Rotating the snapshot key to server and adding a delegation ----\n")
        self.client.run(["key", "rotate", self.repo_name, "snapshot", "-r"], self.dir)
        self.client.run(
            ["delegation", "add", self.repo_name, "targets/releases",
             os.path.join(reporoot(), "fixtures", "secure.example.com.crt"), "--all-paths"],
            self.dir)
        self.client.run(["publish", self.repo_name], self.dir)

        print("---- Listing delegations ----\n")
        delegations1 = self.client.run(["delegation", "list", self.repo_name], self.dir)
        delegations2 = self.client.run(["delegation", "list", self.repo_name], tempdir)

        assert delegations1 == delegations2, "delegation lists not equal: \n{0}\n{1}".format(
            delegations1, delegations2)
        assert "targets/releases" in delegations1, "targets/releases delegation not added"

        # add key to tempdir, publish target
        print("---- Publishing a target using a delegation ----\n")
        self.client.run(
            ["key", "import", os.path.join(reporoot(), "fixtures", "secure.example.com.key"),
             "-r", "targets/releases"],
            tempdir)
        self.client.run(
            ["add", self.repo_name, "add_delegation_test", tempfile, "-r", "targets/releases"],
            tempdir)
        self.client.run(["publish", self.repo_name], tempdir)

        print("---- Listing and validating delegation repo test targets ----\n")
        targets1 = self.client.run(["list", self.repo_name], self.dir)
        targets2 = self.client.run(["list", self.repo_name], tempdir)

        assert targets1 == targets2, "targets lists not equal: \n{0}\n{1}".format(
            targets1, targets2)
        expected_target = [line for line in targets1.split("\n")
                           if line.strip().startswith("add_delegation_test") and
                           line.strip().endswith("targets/releases")]
        assert len(expected_target) == 1, "could not find target added to targets/releases"

    def root_rotation_test(self, tempfile, tempdir):
        """
        Test root rotation
        """
        print("---- Figuring out what the old keys are  ----\n")

        # update the tempdir
        self.client.run(["list", self.repo_name], tempdir)

        output = self.client.run(["key", "list"], self.dir)
        orig_root_key_info = [line.strip() for line in output.split("\n")
                              if line.strip().startswith('root')]
        assert len(orig_root_key_info) == 1

        # this should be replaced with notary info later
        with open(os.path.join(tempdir, "tuf", self.repo_name, "metadata", "root.json")) as root:
            root_json = json.load(root)
            old_root_num_keys = len(root_json["signed"]["keys"])
            old_root_certs = root_json["signed"]["roles"]["root"]["keyids"]
            assert len(old_root_certs) == 1

        print("---- Rotating root key  ----\n")
        # rotate root, check that we have a new key - this is interactive, so pass input
        self.client.run(["key", "rotate", self.repo_name, "root"], self.dir, stdinput="yes")
        output = self.client.run(["key", "list"], self.dir)
        new_root_key_info = [line.strip() for line in output.split("\n")
                             if line.strip().startswith('root') and
                             line.strip() != orig_root_key_info[0]]
        assert len(new_root_key_info) == 1

        # update temp dir and make sure we can download the update
        self.client.run(["list", self.repo_name], tempdir)
        with open(os.path.join(tempdir, "tuf", self.repo_name, "metadata", "root.json")) as root:
            root_json = json.load(root)
            assert len(root_json["signed"]["keys"]) == old_root_num_keys + 1, (
                "expected {0} base keys, but got {1}".format(
                    old_root_num_keys + 1, len(root_json["signed"]["keys"])))

            root_certs = root_json["signed"]["roles"]["root"]["keyids"]

            assert len(root_certs) == 1, "expected 1 valid root key, got {0}".format(
                len(root_certs))
            assert root_certs != old_root_certs, "root key has not been rotated"

        print("---- Ensuring we can still publish  ----\n")
        # make sure we can still publish from both repos
        self.client.run(
            ["key", "import", os.path.join(reporoot(), "fixtures", "secure.example.com.key"),
             "-r", "targets/releases"],
            tempdir)
        self.client.run(
            ["add", self.repo_name, "root_rotation_test_delegation_add", tempfile,
             "-r", "targets/releases"],
            tempdir)
        self.client.run(["publish", self.repo_name], tempdir)
        self.client.run(["add", self.repo_name, "root_rotation_test_targets_add", tempfile],
                        self.dir)
        self.client.run(["publish", self.repo_name], self.dir)

        targets1 = self.client.run(["list", self.repo_name], self.dir)
        targets2 = self.client.run(["list", self.repo_name], tempdir)

        assert targets1 == targets2, "targets lists not equal: \n{0}\n{1}".format(
            targets1, targets2)

        lines = [line.strip() for line in targets1.split("\n")]
        expected_targets = [
            line for line in lines
            if (line.startswith("root_rotation_test_delegation_add") and
                line.endswith("targets/releases"))
            or (line.startswith("root_rotation_test_targets_add") and line.endswith("targets"))]
        assert len(expected_targets) == 2

    def run(self):
        """
        Run tests
        """
        for test_func in (self.basic_repo_test, self.add_delegation_test, self.root_rotation_test):
            _, tempfile = mkstemp()
            with open(tempfile, 'wb') as handle:
                handle.write(test_func.__name__ + "\n")

            tempdir = mkdtemp(suffix="_temp")

            try:
                test_func(tempfile, tempdir)
            except Exception:
                raise
            else:
                cleanup(tempfile, tempdir)

        cleanup(self.dir)

def wait_for_server(server, timeout_in_seconds):
    """
    Attempts to contact the server until it is up
    """
    command = ["curl", server]
    if server is None:
        server = "https://notary-server:4443"
        command = ["curl", "--cacert", os.path.join(reporoot(), "fixtures", "root-ca.crt"),
                   server + "/_notary_server/health"]

    start = time()
    succeeded = 0
    while time() <= start + timeout_in_seconds:
        proc = Popen(command, stderr=PIPE, stdin=PIPE)
        proc.communicate()
        if proc.poll():
            sleep(11)
            continue
        else:
            succeeded += 1
            if succeeded > 1:
                break

    if succeeded < 2:
        raise Exception(
            "Could not connect to {0} after {1} seconds.".format(server, timeout_in_seconds))

    # sleep for 30 extra seconds to wait for the server to connect to the signer
    sleep(30)


def run():
    """
    Run the client tests
    """
    repo_name, server, username = parse_args()
    if not repo_name:
        repo_name = uuid4().hex
    if server is not None:
        server = server.lower().strip()

    if server in ("https://notary-server:4443", "https://notaryserver:4443", ""):
        server = None

    print("building a new client binary")
    call(['make', '-C', reporoot(), 'client'])
    print('---')

    username_passwd = ()
    if username is not None and username.strip():
        username = username.strip()
        password = getpass("password to server for user {0}: ".format(username))
        username_passwd = (username, password)

    wait_for_server(server, 120)

    Tester(repo_name, Client(server, username_passwd)).run()

    try:
        with open("/test_output/SUCCESS", 'wb') as successFile:
            successFile.write("OK")
        os.chmod("/test_output/SUCCESS", 0o777)
    except IOError:
        pass

if __name__ == "__main__":
    run()
