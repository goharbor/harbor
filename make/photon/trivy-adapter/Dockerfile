ARG golang_image
ARG harbor_base_image_version
ARG harbor_base_namespace

# Build the Trivy scanner from source instead of downloading its GitHub release
# tarball. Aqua prunes and yanks old release assets, so the pinned download URL
# 404s once a version ages out, which permanently breaks build-from-source of
# frozen Harbor tags. `go install <pkg>@<version>` resolves the module from the
# Go module proxy, which serves checksummed source archives (verified against
# sum.golang.org) and keeps them after the GitHub release is gone. Set
# GOPROXY=direct to resolve straight from the git tag instead.
FROM ${golang_image} AS trivy-builder

ARG trivy_version
ARG TARGETARCH=amd64

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${TARGETARCH}
# Set by Trivy's own release build, see .goreleaser.yml at the pinned tag.
ENV GOEXPERIMENT=jsonv2

# The flags mirror Trivy's .goreleaser.yml: a static, stripped binary with
# app.ver stamped without the leading "v", so `trivy --version` reports e.g.
# 0.71.1. Both the ldflags path and the build env have moved between Trivy
# releases, so re-check .goreleaser.yml at the new tag when bumping
# TRIVYVERSION.
RUN GOBIN=/out go install \
    -ldflags "-s -w -extldflags '-static' -X github.com/aquasecurity/trivy/pkg/version/app.ver=${trivy_version#v}" \
    github.com/aquasecurity/trivy/cmd/trivy@${trivy_version}

FROM ${harbor_base_namespace}/harbor-trivy-adapter-base:${harbor_base_image_version}

ARG trivy_version

COPY ./make/photon/common/install_cert.sh /home/scanner
COPY ./make/photon/trivy-adapter/entrypoint.sh /home/scanner
COPY --from=trivy-builder /out/trivy /usr/local/bin/trivy
COPY ./make/photon/trivy-adapter/binary/scanner-trivy /home/scanner/bin/scanner-trivy


RUN chown -R scanner:scanner /etc/pki/tls/certs \
    && chown scanner:scanner /home/scanner/entrypoint.sh && chmod u+x /home/scanner/entrypoint.sh \
    && chown scanner:scanner /usr/local/bin/trivy && chmod u+x /usr/local/bin/trivy \
    && chown scanner:scanner /home/scanner/bin/scanner-trivy && chmod u+x /home/scanner/bin/scanner-trivy \
    && chown scanner:scanner /home/scanner/install_cert.sh && chmod u+x /home/scanner/install_cert.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl --fail -s http://localhost:8080/probe/healthy || curl -k --fail -s https://localhost:8443/probe/healthy || exit 1

ENV TRIVY_VERSION=${trivy_version}

USER scanner

ENTRYPOINT ["/home/scanner/entrypoint.sh"]
