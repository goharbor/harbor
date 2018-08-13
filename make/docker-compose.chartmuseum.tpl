version: '2'
services:
  ui:
    networks:
      harbor-chartmuseum:
        aliases:
          - harbor-ui
  redis:
    networks:
      harbor-chartmuseum:
        aliases:
          - redis
  chartmuseum:
    container_name: chartmuseum
    image: goharbor/chartmuseum-photon:__chartmuseum_version__
    restart: always
    networks:
      - harbor-chartmuseum
    depends_on:
      - redis
    volumes:
      - /data/chart_storage:/chart_storage:z
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "chartmuseum"
    env_file:
      ./common/config/chartserver/env
networks:
  harbor-chartmuseum:
    external: false
