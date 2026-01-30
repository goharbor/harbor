

cd ~/go/src/github.com/goharbor/harbor
sudo make install
sudo make compile_portal
sudo /usr/bin/docker build --no-cache --network=default --pull=true --build-arg harbor_base_image_version=dev --build-arg harbor_base_namespace=goharbor --build-arg NODE=node:16.18.0 --build-arg npm_registry=https://registry.npmjs.org -f /home/max/go/src/github.com/goharbor/harbor/make/photon/portal/Dockerfile -t goharbor/harbor-portal:dev .
sudo docker compose -f /home/max/go/src/github.com/goharbor/harbor/make/docker-compose.yml up -d



cd ~/go/src/github.com/goharbor/harbor/src/portal
npm install
npm run start
npm run start:debug


#  goharbor/harbor-db:dev
sudo docker exec -i -t harbor-db sh
psql -U postgres -d registry


tail -f /var/log/harbor/core.log 



############## Errors ###############################
Handler crashed with error <Ormer.QueryTable> 
table name: `github.com/goharbor/harbor/src/pkg/role/model.Role` not exists


