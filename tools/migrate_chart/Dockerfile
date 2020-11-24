FROM python:3.8.5-slim

ENV HELM_EXPERIMENTAL_OCI=1
ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

COPY ./migrate_chart.py ./migrate_chart.sh /
ADD https://get.helm.sh/helm-v3.2.4-linux-amd64.tar.gz /

RUN tar zxvf /helm-v3.2.4-linux-amd64.tar.gz && \
    pip install click==7.1.2 && \
    pip install requests==2.24.0 && \
    chmod +x /migrate_chart.sh ./migrate_chart.py

ENTRYPOINT [ "/migrate_chart.py" ]