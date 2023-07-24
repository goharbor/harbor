ARG GOLANG
FROM ${GOLANG}

RUN tdnf install libstdc++ -y

ARG SPECTRAL_VERSION
RUN curl -fsSL -o /usr/bin/spectral https://github.com/stoplightio/spectral/releases/download/$SPECTRAL_VERSION/spectral-linux && chmod +x /usr/bin/spectral

ENTRYPOINT ["/usr/bin/spectral"]
CMD ["--version"]
