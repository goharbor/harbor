FROM reg-bj.eng.vmware.com/sharedrepo/harbor-clarity-base:0.8.0


COPY make/dev/nodeclarity/clarity-seed /clarity-seed
COPY make/dev/nodeclarity/index.html /clarity-seed
COPY make/dev/nodeclarity/entrypoint.sh /clarity-seed

WORKDIR /clarity-seed

RUN chmod u+x entrypoint.sh

VOLUME ["/clarity-seed/src/app", "/clarity-seed/dist"]

ENTRYPOINT ["/clarity-seed/entrypoint.sh"]
