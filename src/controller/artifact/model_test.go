package artifact

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalJSONWithACC(t *testing.T) {
	data := []byte(`[{"accessories":[{"artifact_id":9,"creation_time":"2022-01-20T09:18:50.993Z","digest":"sha256:a7caa2636af890178a0b8c4cdbc47ced4dbdf29a1680e9e50823e85ce35b28d3","icon":"","id":4,"size":501,"subject_artifact_id":8,"type":"signature.cosign"}],
	"addition_links":{"build_history":{"absolute":false,"href":"/api/v2.0/projects/source_project011642670285/repositories/redis/artifacts/sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c/additions/build_history"},
	"vulnerabilities":{"absolute":false,"href":"/api/v2.0/projects/source_project011642670285/repositories/redis/artifacts/sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c/additions/vulnerabilities"}},
	"digest":"sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c",
	"extra_attrs":{"architecture":"amd64","author":"","config":{"Cmd":["redis-server"],"Entrypoint":["docker-entrypoint.sh"],"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","GOSU_VERSION=1.11","REDIS_VERSION=5.0.7","REDIS_DOWNLOAD_URL=redis-5.0.7.tar.gz","REDIS_DOWNLOAD_SHA=61db74eabf6801f057fd24b590"],"ExposedPorts":{"6379/tcp":{}},"Volumes":{"/data":{}},"WorkingDir":"/data"},"created":"2020-01-03T01:29:15.570681619Z","os":"linux"},"icon":"sha256:0048162a053eef4d4ce3fe7518615bef084403614f8bca43b40ae2e762e11e06","id":8,"labels":null,"manifest_media_type":"application/vnd.docker.distribution.manifest.v2+json","media_type":"application/vnd.docker.container.image.v1+json","project_id":7,"pull_time":"2022-01-20T09:18:50.783Z","push_time":"2022-01-20T09:18:50.290Z","references":null,"repository_id":5,"size":35804754,
	"tags":[{"artifact_id":8,"id":6,"immutable":false,"name":"latest","pull_time":"2022-01-20T09:18:50.783Z","push_time":"2022-01-20T09:18:50.303Z","repository_id":5,"signed":false}],"type":"IMAGE"}]`)

	var artifact []Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fail()
	}

	assert.Equal(t, int64(9), artifact[0].Accessories[0].GetData().ArtifactID)
	assert.Equal(t, "latest", artifact[0].Tags[0].Name)
	assert.Equal(t, "amd64", artifact[0].ExtraAttrs["architecture"])

	_, ok := artifact[0].Accessories[0].(*cosign.Signature)
	assert.True(t, ok)
}

func TestUnmarshalJSONWithACCPartial(t *testing.T) {
	data := []byte(`[{"accessories":[{"artifact_id":9,"creation_time":"2022-01-20T09:18:50.993Z","digest":"sha256:a7caa2636af890178a0b8c4cdbc47ced4dbdf29a1680e9e50823e85ce35b28d3","icon":"","id":4,"size":501,"subject_artifact_id":8,"type":"signature.cosign"}, {"artifact_id":2, "type":"signature.cosign"}],
	"digest":"sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c","tags":[{"artifact_id":8,"id":6,"immutable":false,"name":"latest","pull_time":"2022-01-20T09:18:50.783Z","push_time":"2022-01-20T09:18:50.303Z","repository_id":5,"signed":false}],"type":"IMAGE"}]`)

	var artifact []Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fail()
	}

	assert.Equal(t, int64(9), artifact[0].Accessories[0].GetData().ArtifactID)
	assert.Equal(t, int64(2), artifact[0].Accessories[1].GetData().ArtifactID)
	assert.Equal(t, "latest", artifact[0].Tags[0].Name)
	_, ok := artifact[0].Accessories[1].(*cosign.Signature)
	assert.True(t, ok)
}

func TestUnmarshalJSONWithACCUnknownType(t *testing.T) {
	data := []byte(`[{"accessories":[{"artifact_id":9,"creation_time":"2022-01-20T09:18:50.993Z","digest":"sha256:a7caa2636af890178a0b8c4cdbc47ced4dbdf29a1680e9e50823e85ce35b28d3","icon":"","id":4,"size":501,"subject_artifact_id":8}],
	"digest":"sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c","tags":[{"artifact_id":8,"id":6,"immutable":false,"name":"latest","pull_time":"2022-01-20T09:18:50.783Z","push_time":"2022-01-20T09:18:50.303Z","repository_id":5,"signed":false}],"type":"IMAGE"}]`)

	var artifact []Artifact
	err := json.Unmarshal(data, &artifact)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "accessory type  not support")
}

func TestUnmarshalJSONWithoutACC(t *testing.T) {
	data := []byte(`[{"addition_links":{"build_history":{"absolute":false,"href":"/api/v2.0/projects/source_project011642670285/repositories/redis/artifacts/sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c/additions/build_history"},
"vulnerabilities":{"absolute":false,"href":"/api/v2.0/projects/source_project011642670285/repositories/redis/artifacts/sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c/additions/vulnerabilities"}},
"digest":"sha256:e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c",
"extra_attrs":{"architecture":"amd64","author":"","config":{"Cmd":["redis-server"],"Entrypoint":["docker-entrypoint.sh"],"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","GOSU_VERSION=1.11","REDIS_VERSION=5.0.7","REDIS_DOWNLOAD_URL=redis-5.0.7.tar.gz","REDIS_DOWNLOAD_SHA=61db74eabf6801f057fd24b590"],"ExposedPorts":{"6379/tcp":{}},"Volumes":{"/data":{}},"WorkingDir":"/data"},"created":"2020-01-03T01:29:15.570681619Z","os":"linux"},"icon":"sha256:0048162a053eef4d4ce3fe7518615bef084403614f8bca43b40ae2e762e11e06","id":8,"labels":null,"manifest_media_type":"application/vnd.docker.distribution.manifest.v2+json","media_type":"application/vnd.docker.container.image.v1+json","project_id":7,"pull_time":"2022-01-20T09:18:50.783Z","push_time":"2022-01-20T09:18:50.290Z","references":null,"repository_id":5,"size":35804754,
"tags":[{"artifact_id":8,"id":6,"immutable":false,"name":"latest","pull_time":"2022-01-20T09:18:50.783Z","push_time":"2022-01-20T09:18:50.303Z","repository_id":5,"signed":false}],"type":"IMAGE"}]`)

	var artifact []Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fail()
	}

	assert.Equal(t, "latest", artifact[0].Tags[0].Name)
	assert.Equal(t, "amd64", artifact[0].ExtraAttrs["architecture"])
}

func TestUnmarshalJSONWithAccNull(t *testing.T) {
	data := []byte(`{"accessories":null,"addition_links":{"build_history":{"absolute":false,"href":"/api/v2.0/projects/project-1643104251947/repositories/test3/artifacts/sha256:fc00fd623137fa47bd5b3f/additions/build_history"}},
"digest":"sha256:fc00fd623137fa47bd5b3f2","extra_attrs":{"architecture":"amd64","author":"","config":{"Cmd":["dd","if=/dev/urandom","of=test","bs=1M","count=1"],"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"]},"created":"2022-01-25T09:51:13.904772229Z","os":"linux"},"icon":"sha256:0048162a053eef4d4ce3fe7518615bef084403614f8bca43b40ae2e762e11e06","id":12,"labels":null,"manifest_media_type":"application/vnd.docker.distribution.manifest.v2+json","media_type":"application/vnd.docker.container.image.v1+json","project_id":8,"pull_time":"0001-01-01T00:00:00.000Z","push_time":"2022-01-25T09:51:14.394Z","references":null,"repository_id":6,"size":1816010,"tags":[{"artifact_id":12,"id":13,"immutable":false,"name":"1.0","pull_time":"0001-01-01T00:00:00.000Z","push_time":"2022-01-25T09:51:14.406Z","repository_id":6,"signed":false}],"type":"IMAGE"}`)

	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fail()
	}

	assert.Equal(t, "1.0", artifact.Tags[0].Name)
	assert.Equal(t, "amd64", artifact.ExtraAttrs["architecture"])
}

func TestUnmarshalJSONWithNull(t *testing.T) {
	data := []byte(`{}`)
	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fail()
	}
	assert.Equal(t, "", artifact.Digest)
}

func TestUnmarshalJSONWithPartial(t *testing.T) {
	data := []byte(`{"digest":"sha256:1234","media_type":"application/vnd.docker.container.image.v1+json","project_id":8,"pull_time":"0001-01-01T00:00:00.000Z","push_time":"2022-01-25T09:51:14.394Z","references":null}`)
	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fail()
	}
	assert.Equal(t, "sha256:1234", artifact.Digest)
	assert.Equal(t, "", artifact.Type)
	assert.Equal(t, "application/vnd.docker.container.image.v1+json", artifact.MediaType)
}
