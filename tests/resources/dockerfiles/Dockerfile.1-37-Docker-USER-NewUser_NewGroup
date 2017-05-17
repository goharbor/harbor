FROM busybox:latest
RUN addgroup newuser && adduser -H -S newuser -G newuser
USER newuser
CMD id
