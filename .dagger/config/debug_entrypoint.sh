#!/bin/sh

# Execute the command (passed as $1 arg)
echo "Executing: dlv --headless=true --listen=0.0.0.0:4001 --accept-multiclient --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log=true --continue --api-version=2 exec $1"

# Start the dlv process in the background
# /root/go/bin/dlv exec --headless --listen localhost:$2 $1
dlv --headless=true --listen=0.0.0.0:4001 --accept-multiclient --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log=true --continue --api-version=2 exec $1
# dlv --headless=true --listen=0.0.0.0:4001 --accept-multiclient --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log=true --api-version=2 attach $pid
