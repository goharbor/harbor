FROM node:7.5.0

ENV http_proxy="http://proxy.vmware.com:3128"
ENV https_proxy="http://proxy.vmware.com:3128"

RUN git clone https://github.com/vmware/clarity-seed.git /clarity-seed

COPY index.html /clarity-seed
COPY entrypoint.sh /clarity-seed

WORKDIR /clarity-seed

RUN npm install -g @angular/cli && \
    npm install && \
    chmod u+x entrypoint.sh

VOLUME ["/clarity-seed/src/app", "/clarity-seed/dist"]

ENTRYPOINT ["/clarity-seed/entrypoint.sh"]
