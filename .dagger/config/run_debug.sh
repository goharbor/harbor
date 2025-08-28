#!/bin/sh

envFile="/envFile"

# Check if the file exists
if [ ! -f "$envFile" ]; then
  echo "Environment file not found: $envFile"
  exit 1
fi

# Read the file and export each line as an environment variable
while IFS='=' read -r key value; do
  # Ignore empty lines or comments
  if [ -z "$key" ] || [ "${key:0:1}" = "#" ]; then
    continue
  fi

  # Export the environment variable
  export "$key"="$value"

  # print the variable to verify it's set
  echo "Set $key=$value"
done < "$envFile"

# Execute the core command directly
# $1 &

# Get process ID of previously ran command
# pid=$!

# Execute the command (passed as $1 arg)
echo "Executing: $1, with pid: $pid, with debug enabled at port: $2"

# Start the dlv process in the background
# /root/go/bin/dlv exec --headless --listen localhost:$2 $1
dlv --headless=true --listen=0.0.0.0:4001 --accept-multiclient --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log=true --continue --api-version=2 exec /core
# /root/go/bin/dlv --headless=true --listen=localhost:4001 --accept-multiclient --log-output=debugger,debuglineerr,gdbwire,lldbout,rpc --log=true --api-version=2 attach $pid
