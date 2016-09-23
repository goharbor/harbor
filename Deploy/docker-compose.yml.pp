version: '2'
services:
  registry:
    image: library/registry:2.5.0
    restart: always
    ports:
      - 5000:5000
    volumes:
      - /data/registry:/storage
      - ./config/registry/:/etc/registry/
    environment:
      - GODEBUG=netdns=cgo
    command:
      ["serve", "/etc/registry/config.yml"]
  ui:
    build:
      context: ../
      dockerfile: Deploy/ui/Dockerfile.pp
    env_file:
      - ./config/ui/env
    restart: always
    ports:
      - 80:80
    volumes:
      - ./config/ui/app.conf:/etc/ui/app.conf
      - ./config/ui/private_key.pem:/etc/ui/private_key.pem
