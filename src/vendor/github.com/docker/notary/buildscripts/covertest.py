#!/usr/bin/env python

"""
This module is used to run tests with full coverage-reports.

It's a way to provide accurate -coverpkg arguments to `go test`.

To run over all packages:
buildscripts/covertest.py --tags "tag1 tag2" --testopts="-v -race"

To run over some packages:
buildscripts/covertest.py --pkgs "X"
"""

from __future__ import print_function

import os
import subprocess
import sys
from argparse import ArgumentParser


class CoverageRunner(object):
    """
    CoverageRunner generates coverage profiles for a packages within a base package,
    making sure to produce numbers for every relevant package that it depends on or that its
    tests depends on.

    Ideally, we can just do `-coverpkg=all`, but (1) that includes all vendored packages,
    and (2) even if we do `-coverpkg=$(go list ./... | grep -v vendor)`, compiling
    the test binary slows down a bit for each package in the `-coverpkg` list.

    This has to calculate the recursive dependencies for the package being tested, which
    is the union of the following sets of dependencies:
    - recursive dependencies of the non-test package: produced with `go list -f {{.Deps}}`
    - test imports (which are not necessarily included with the previous go list command):
      produced with `go list -f {{.TestImports}}`
    - test imports from a test in a different package, such as from package `<pkg>_test`:
      produced with `go list -f {{.XTestImports}}`
    """
    def __init__(self, buildtags):
        self.base_pkg = subprocess.check_output("go list".split()).strip()
        self.buildtags = buildtags
        self.tag_args = ()
        if buildtags != "":
            self.tag_args = ("-tags", self.buildtags)

        self.recursive_pkg_deps, self.test_imports = self._get_all_pkg_info()

    def _filter_pkgs(self, pkgs):
        """
        Returns a filtered copy of the list that only contains source packages derived from the
        base package, minus any vendored packages.
        """
        pkgs = [pkg.strip() for pkg in pkgs]
        return [
            pkg for pkg in pkgs
            if pkg.startswith(self.base_pkg) and not pkg.startswith(os.path.join(self.base_pkg, "vendor/"))
        ]

    def _go_list(self, *args):
        """
        Runs go list with some extra formatting and args
        """
        return subprocess.check_output(("go", "list") + self.tag_args + args).strip().split("\n")

    def _get_all_pkg_info(self):
        """
        Returns all dependency and test info for every package
        """
        all_pkgs = self._filter_pkgs(self._go_list("./..."))
        # for every package, list the deps, the test files, the test imports, and the external package test imports
        big_list = self._go_list(
            "-f", "{{.ImportPath}}:{{.Deps}}:{{.TestImports}}:{{.XTestImports}}", *all_pkgs)
        recursive_deps = {}
        test_imports = {}

        for line in big_list:
            tokens = [token.strip().lstrip('[').rstrip(']').strip() for token in line.split(":", 3)]
            pkg = tokens[0].strip()

            recursive_deps[pkg] = set(self._filter_pkgs(tokens[1].split() + [pkg]))
            if tokens[2] or tokens[3]:
                test_imports[pkg] = set(
                    self._filter_pkgs(tokens[2].split()) + self._filter_pkgs(tokens[3].split()))

        return recursive_deps, test_imports

    def get_coverprofile_filename(self, pkg):
        """
        Returns the cover profile that should be produced for a package.
        """
        items = ("coverage", "txt")
        if self.buildtags:
            items = ("coverage", self.buildtags.replace(",", ".").replace(" ", ""), "txt")
        return os.path.join(pkg.replace(self.base_pkg + "/", ""), ".".join(items))

    def get_pkg_recursive_deps(self, pkg):
        """
        Returns all package dependencies for which coverage should be generated - this includes the actual
        package recursive dependencies, as well as the recursive test dependencies for that package.
        """
        return self.recursive_pkg_deps[pkg].union(
            *[self.recursive_pkg_deps[test_import] for test_import in self.test_imports.get(pkg, ())])

    def run(self, pkgs=(), testopts="", covermode="atomic", debug=False):
        """
        Run go test with coverage over the the given packages, with the following given options
        """
        pkgs = pkgs or self.test_imports.keys()
        pkgs.sort()

        cmds = []

        if debug:
            print("`go test` commands that will be run:\n")

        for pkg in pkgs:
            pkg_deps = self.get_pkg_recursive_deps(pkg)
            cmd = ["go", "test"] + list(self.tag_args)
            if pkg in self.test_imports:
                cmd += testopts.split() + [
                        "-covermode", covermode,
                        "-coverprofile", self.get_coverprofile_filename(pkg),
                        "-coverpkg", ",".join(pkg_deps)]
            cmd += [pkg]
            if debug:
                print("\t" + " ".join(cmd))
            cmds.append((cmd, pkg_deps))

        if debug:
            print("\nResults:\n")

        for cmd, pkg_deps in cmds:
            app = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
            for line in app.stdout:
                # we are trying to generate coverage reports for everything in the base package, and some may not be
                # actually exercised in the test.  So ignore this particular warning.
                if not line.startswith("warning: no packages being tested depend on {0}".format(self.base_pkg)):
                    line = line.replace("statements in " + ", ".join(pkg_deps),
                                        "statements in <{0} dependencies>".format(len(pkg_deps)))
                    sys.stdout.write(line)

            app.wait()
            if app.returncode != 0:
                print("\n\nTests failed.\n")
                sys.exit(app.returncode)

    def print_test_deps_not_in_package_deps(self):
        """
        Print out any packages for which there are test dependencies not returned by {{.Deps}}
        """
        extras = []
        for key, rec_deps in self.recursive_pkg_deps.items():
            any = self.test_imports.get(key, set()).difference(rec_deps, set([key]))
            if any:
                extras.append((key, any))

        if extras:
            print("Packages whose tests have extra dependencies not listed in `go list -f {{.Deps}}`:")
            for pkg, deps in extras:
                print("\t{0}: {1}".format(pkg, ", ".join(deps)))
            print("\n")


def parseArgs(args=None):
    """
    CLI option parsing
    """
    parser = ArgumentParser()
    parser.add_argument("--testopts", help="Options to pass for testing, such as -race or -v", default="")
    parser.add_argument("--pkgs", help="Packages to test specifically, otherwise we test all the packages", default="")
    parser.add_argument("--tags", help="Build tags to pass to go", default="")
    parser.add_argument("--debug", help="Print extra debugging info regarding the coverage run", action="store_true")
    return parser.parse_args(args)

if __name__ == "__main__":
    args = parseArgs()
    pkgs = args.pkgs.strip().split()

    runner = CoverageRunner(args.tags)
    if args.debug:
        runner.print_test_deps_not_in_package_deps()
    runner.run(pkgs=pkgs, testopts=args.testopts, debug=args.debug)
