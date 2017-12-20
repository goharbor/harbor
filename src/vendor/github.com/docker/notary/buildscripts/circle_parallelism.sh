#!/usr/bin/env bash

set -e
case $CIRCLE_NODE_INDEX in
0) docker run --rm -e NOTARY_BUILDTAGS=pkcs11 --env-file buildscripts/env.list --user notary notary_client bash -c "make ci && codecov"
   ;;
1) docker run --rm -e NOTARY_BUILDTAGS=none --env-file buildscripts/env.list --user notary notary_client bash -c "make ci && codecov"
   ;;
2) SKIPENVCHECK=1 make TESTDB=mysql testdb
   SKIPENVCHECK=1 make TESTDB=mysql integration
   SKIPENVCHECK=1 make cross  # just trying not to exceed 5 builders
   ;;
3) SKIPENVCHECK=1 make TESTDB=rethink testdb
   SKIPENVCHECK=1 make TESTDB=rethink integration
   ;;
4) docker run --rm -e NOTARY_BUILDTAGS=pkcs11 notary_client make lint
   SKIPENVCHECK=1 make TESTDB=postgresql testdb
   SKIPENVCHECK=1 make TESTDB=postgresql integration
   ;;
esac
