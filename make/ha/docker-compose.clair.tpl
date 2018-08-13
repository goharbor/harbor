version: '2'
services:
  ui:
    networks:
      harbor-clair:
        aliases:
          - harbor-ui
  jobservice:
    networks:
      - harbor-clair
  registry:
    networks:
      - harbor-clair
  clair:
    networks:
      - harbor-clair
    container_name: clair
    image: goharbor/clair-photon:__clair_version__
    restart: always
    cpu_quota: 150000
    depends_on:
      - log
    volumes:
      - ./common/config/clair:/config
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "clair"
    env_file:
      ./common/config/clair/clair_env
networks:
  harbor-clair:
    external: false
