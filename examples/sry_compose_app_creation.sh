#!/bin/bash

echo "http --auth admin:Harbor12345 post localhost:8080/api/v3/repositories/library/drone/apps @./sry_compose_app_creation.json"
http --auth admin:Harbor12345 post localhost:8080/api/v3/repositories/library/drone/apps @./sry_compose_app_creation.json
