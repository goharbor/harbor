version: '2'
services:
  core:
    networks:
      - harbor-notary
  proxy:
    networks:
      - harbor-notary
  postgresql:
    networks:
      harbor-notary:
        aliases:
          - harbor-db
  notary-server:
    image: goharbor/notary-server-photon:__notary_version__
    container_name: notary-server
    restart: always
    networks:
      - notary-sig
      - harbor-notary
    dns_search: .
    volumes:
      - ./common/config/notary:/etc/notary:z
    env_file:
      - ./common/config/notary/server_env
    depends_on:
      - postgresql
      - notary-signer
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "notary-server"
  notary-signer:
    image: goharbor/notary-signer-photon:__notary_version__
    container_name: notary-signer
    restart: always
    networks:
      harbor-notary:
      notary-sig:
        aliases:
          - notarysigner
    dns_search: .
    volumes:
      - ./common/config/notary:/etc/notary:z
    env_file:
      - ./common/config/notary/signer_env
    depends_on:
      - postgresql
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "notary-signer"
networks:
  harbor-notary:
    external: false
  notary-sig:
    external: false