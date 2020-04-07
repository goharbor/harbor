ARG harbor_base_image_version
FROM goharbor/harbor-clair-adapter-base:${harbor_base_image_version}

COPY ./make/photon/common/install_cert.sh /home/clair-adapter
COPY ./make/photon/clair-adapter/entrypoint.sh /home/clair-adapter
COPY ./make/photon/clair-adapter/binary/harbor-scanner-clair /clair-adapter/clair-adapter

RUN chown -R clair-adapter:clair-adapter /etc/pki/tls/certs \
    && chown -R clair-adapter:clair-adapter /clair-adapter && chmod u+x /clair-adapter/clair-adapter \
    && chown clair-adapter:clair-adapter /home/clair-adapter/entrypoint.sh && chmod u+x /home/clair-adapter/entrypoint.sh \
    && chown clair-adapter:clair-adapter /home/clair-adapter/install_cert.sh && chmod u+x /home/clair-adapter/install_cert.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS http://127.0.0.1:8080/probe/healthy || curl -k -sS https://127.0.0.1:8443/probe/healthy || exit 1

USER clair-adapter

ENTRYPOINT ["/home/clair-adapter/entrypoint.sh"]