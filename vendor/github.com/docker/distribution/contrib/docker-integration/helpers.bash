# Start docker daemon
function start_daemon() {
	# Drivers to use for Docker engines the tests are going to create.
	STORAGE_DRIVER=${STORAGE_DRIVER:-overlay}
	EXEC_DRIVER=${EXEC_DRIVER:-native}

	docker --daemon --log-level=panic \
		--storage-driver="$STORAGE_DRIVER" --exec-driver="$EXEC_DRIVER" &
	DOCKER_PID=$!

	# Wait for it to become reachable.
	tries=10
	until docker version &> /dev/null; do
		(( tries-- ))
		if [ $tries -le 0 ]; then
			echo >&2 "error: daemon failed to start"
			exit 1
		fi
		sleep 1
	done
}
