#!/bin/bash

USERNAME=admin
PASSWORD=Dataman1234
#HARBOR_URL=http://forward.dataman-inc.com/
HARBOR_URL=devregistry.dataman-inc.com:5005

PROJECT_NAME=library
REPO_NAME=nginx

DESCRIPTION=最流行的web服务器
CATEGORY=中间件
ISPUBLIC=1


# 更新分类
update_repo_category()
{
  echo ""
  echo " ====  UPDATING $1's category to $2 ===="
  echo ""
  curl -u $USERNAME:$PASSWORD -X PUT -k -H "Content-Type: application/json" $HARBOR_URL/api/v3/repositories/$PROJECT_NAME/$1 -d " { \"category\": \"$2\" } "
}

# 更新描述
update_repo_description()
{
  echo ""
  echo " ====  UPDATING $1's description to $2 ===="
  echo ""
  curl -u $USERNAME:$PASSWORD -X PUT -k -H "Content-Type: application/json"  $HARBOR_URL/api/v3/repositories/$PROJECT_NAME/$1 -d " { \"description\": \"$2\" } "
}

# 更新状态是否可见
update_repo_is_public()
{
  echo ""
  echo " ====  UPDATING $1's publicity to $2 ===="
  echo ""
  curl -u $USERNAME:$PASSWORD -X PUT -k -H "Content-Type: application/json"  $HARBOR_URL/api/v3/repositories/$PROJECT_NAME/$1 -d " { \"isPublic\": $2 } "
}

update_repo_category $REPO_NAME $CATEGORY
update_repo_description $REPO_NAME $DESCRIPTION
update_repo_is_public $REPO_NAME $ISPUBLIC

