version: '2'
services:
  ui:
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
    image: vmware/notary-server-photon:__notary_version__
    container_name: notary-server
    restart: always
    networks:
      - notary-sig
      - harbor-notary
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
    image: vmware/notary-signer-photon:__notary_version__
    container_name: notary-signer
    restart: always
    networks:
      harbor-notary:
      notary-sig:
        aliases:
          - notarysigner
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