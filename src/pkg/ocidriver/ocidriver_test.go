package ocidriver

import (
	"bytes"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testSuite struct {
	suite.Suite
	mockRegistry *httptest.Server
}

func (t *testSuite) TestCase() {
	reg, err := NewRegistry(t.mockRegistry.URL, nil)
	t.Require().Nil(err)
	t.Require().NotNil(reg)

	t.Require().Nil(reg.Ping())

	repos, err := reg.Catalog()
	t.Require().Nil(err)
	t.Require().True(len(repos) > 0)

	// Pull
	repoBusybox, err := reg.GetRepository("library/busybox")
	t.Require().Nil(err)
	t.Require().NotNil(repoBusybox)

	tags, err := repoBusybox.List()
	t.Require().Nil(err)
	t.Require().True(len(tags) > 0)
	t.Require().EqualValues("latest", tags[0].Name)

	isList, manifest, manifestList, manifestv1, err := repoBusybox.GetImageByTag("latest")
	t.Require().Nil(err)
	t.Require().Nil(manifestv1)
	t.Require().True(isList)
	t.Require().Nil(manifest)
	t.Require().NotNil(manifestList)
	t.Require().True(len(manifestList.Manifests) > 0)

	t.Require().EqualValues("amd64", manifestList.Manifests[0].Platform.Architecture)
	t.Require().EqualValues("linux", manifestList.Manifests[0].Platform.OS)
	isList, manifest, manifestList, manifestv1, err = manifestList.Get(0)
	t.Require().Nil(err)
	t.Require().Nil(manifestv1)
	t.Require().False(isList)
	t.Require().Nil(manifestList)
	t.Require().NotNil(manifest)

	t.Require().True(len(manifest.Layers) > 0)
	size, reader, err := repoBusybox.GetBlob(manifest.Layers[0].Digest)
	t.Require().Nil(err)
	t.Require().True(size > 0)
	t.Require().NotNil(reader)

	b, err := ioutil.ReadAll(reader)
	t.Require().Nil(err)
	t.Require().True(strings.HasPrefix(string(b), "binary"), fmt.Sprintf("bad blob:%s", string(b)))

	// Push
	repoUpload, err := reg.GetRepository("library/testupload")
	t.Require().Nil(err)
	d, err := repoUpload.PushBlob(bytes.NewReader([]byte("binary test data")))
	t.Require().Nil(err)
	t.Require().Equal("sha256:8b190e7dfa3dca5a46811ab688ca0b2f83863a099f86800a88a0eb8b117ac5fe", d.Digest.String())

	configstr := `
{
	"schemaVersion": 2,
	"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	"config": {
		"mediaType": "application/vnd.docker.container.image.v1+json",
		"size": 1497,
		"digest": "sha256:6d5fcfe5ff170471fcc3c8b47631d6d71202a1fd44cf3c147e50c8de21cf0648"
	},
	"layers": [{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 760984,
		"digest": "sha256:bdbbaa22dec6b7fe23106d2c1b1f43d9598cd8fc33706cc27c1d938ecd5bffc7"
	}]
}`
	dgst, err := repoUpload.PushManifest("latest", "application/vnd.docker.distribution.manifest.v2+json", []byte(configstr))
	t.Require().Nil(err)
	t.Require().NotNil(dgst)

	err = repoUpload.DeleteManifest(digest.FromString(dgst.String()))
	t.Require().Nil(err)
}

func mockRegistry() (*httptest.Server, error) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/_catalog",
			Handler: getJSONHandler(
				"Content-Type: application/json",
				`
{
	"repositories": [
	"busybox",
	"ubuntu"
	]
}
`),
		},
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: fmt.Sprintf("/v2/library/busybox/tags/list"),
			Handler: getJSONHandler(
				`Content-Type: application/json`,
				`
{
  "name": "busybox",
  "tags": [
    "latest",
    "test"
  ]
}`,
			)},
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: fmt.Sprintf("/v2/library/busybox/manifests/latest"),
			Handler: getJSONHandler(`
Content-Type: application/vnd.docker.distribution.manifest.list.v2+json
Docker-Content-Digest: sha256:6915be4043561d64e0ab0f8f098dc2ac48e077fe23f488ac24b665166898115a
Docker-Distribution-Api-Version: registry/2.0
Etag: "sha256:6915be4043561d64e0ab0f8f098dc2ac48e077fe23f488ac24b665166898115a"`,
				`
{
	"manifests": [{
		"digest": "sha256:edafc0a0fb057813850d1ba44014914ca02d671ae247107ca70c94db686e7de6",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "amd64",
			"os": "linux"
		},
		"size": 527
	}, {
		"digest": "sha256:b19898c529964a48ce66923a06ddbde9cd5ae2472a42891b55cc00f7a60ba676",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "arm",
			"os": "linux",
			"variant": "v5"
		},
		"size": 527
	}, {
		"digest": "sha256:5b752a1ceda0bef058eba99a4b6bd74943dcb69b041dae6027601630b2895bd0",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "arm",
			"os": "linux",
			"variant": "v6"
		},
		"size": 527
	}, {
		"digest": "sha256:7044f6fc222ac87449d87e041eae6b5254012a8b4cbbc35e5b317ac61aa12557",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "arm",
			"os": "linux",
			"variant": "v7"
		},
		"size": 527
	}, {
		"digest": "sha256:50edf1d080946c6a76989d1c3b0e753b62f7d9b5f5e66e88bef23ebbd1e9709c",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "arm64",
			"os": "linux",
			"variant": "v8"
		},
		"size": 527
	}, {
		"digest": "sha256:5987b1ad0938dccfe82bfee4869f0fc0d7cfcabee5fcf08a9ee01ca4844b3bab",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "386",
			"os": "linux"
		},
		"size": 527
	}, {
		"digest": "sha256:8cdf838ff4bfd927345d7f171ea160be0d09895b883f3f7f21fb69fe24be9f10",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "ppc64le",
			"os": "linux"
		},
		"size": 528
	}, {
		"digest": "sha256:b064046f9f0ef7eccc1a1ecaef480d11015b62a3bfaf40a886b8104412ccf698",
		"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
		"platform": {
			"architecture": "s390x",
			"os": "linux"
		},
		"size": 528
	}],
	"mediaType": "application\/vnd.docker.distribution.manifest.list.v2+json",
	"schemaVersion": 2
}`),
		},
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: fmt.Sprintf("/v2/library/busybox/manifests/sha256:edafc0a0fb057813850d1ba44014914ca02d671ae247107ca70c94db686e7de6"),
			Handler: getJSONHandler(`
Content-Type: application/vnd.docker.distribution.manifest.v2+json
Docker-Content-Digest: sha256:edafc0a0fb057813850d1ba44014914ca02d671ae247107ca70c94db686e7de6
Docker-Distribution-Api-Version: registry/2.0
Etag: "sha256:edafc0a0fb057813850d1ba44014914ca02d671ae247107ca70c94db686e7de6"
Date: Thu, 02 Jan 2020 07:08:01 GMT
Strict-Transport-Security: max-age=31536000`,
				`
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1497,
      "digest": "sha256:6d5fcfe5ff170471fcc3c8b47631d6d71202a1fd44cf3c147e50c8de21cf0648"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 760984,
         "digest": "sha256:bdbbaa22dec6b7fe23106d2c1b1f43d9598cd8fc33706cc27c1d938ecd5bffc7"
      }
   ]
}`),
		},
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/library/busybox/blobs/sha256:6d5fcfe5ff170471fcc3c8b47631d6d71202a1fd44cf3c147e50c8de21cf0648", // config blob
			Handler: getJSONHandler(`
Content-Type: application/octet-stream
Content-Length: 1497
Accept-Ranges: bytes
Age: 551367
Cache-Control: public, max-age=14400
ETag: "80684cb8e1f9283f685e7725ad620308"`,
				`
{
	"architecture": "amd64",
	"config": {
		"Hostname": "",
		"Domainname": "",
		"User": "",
		"AttachStdin": false,
		"AttachStdout": false,
		"AttachStderr": false,
		"Tty": false,
		"OpenStdin": false,
		"StdinOnce": false,
		"Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
		"Cmd": ["sh"],
		"ArgsEscaped": true,
		"Image": "sha256:eedc135aa40055799748875f21cbfa13704fb1e1675c6eb3b997ea25a4bf10c4",
		"Volumes": null,
		"WorkingDir": "",
		"Entrypoint": null,
		"OnBuild": null,
		"Labels": null
	},
	"container": "5fd9b79f6aa666515be39ca8e79731094d7db277650d0271346eaa37ea1e3252",
	"container_config": {
		"Hostname": "5fd9b79f6aa6",
		"Domainname": "",
		"User": "",
		"AttachStdin": false,
		"AttachStdout": false,
		"AttachStderr": false,
		"Tty": false,
		"OpenStdin": false,
		"StdinOnce": false,
		"Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
		"Cmd": ["/bin/sh", "-c", "#(nop) ", "CMD [\"sh\"]"],
		"ArgsEscaped": true,
		"Image": "sha256:eedc135aa40055799748875f21cbfa13704fb1e1675c6eb3b997ea25a4bf10c4",
		"Volumes": null,
		"WorkingDir": "",
		"Entrypoint": null,
		"OnBuild": null,
		"Labels": {}
	},
	"created": "2019-12-26T21:20:11.581830492Z",
	"docker_version": "18.06.1-ce",
	"history": [{
		"created": "2019-12-26T21:20:11.445765847Z",
		"created_by": "/bin/sh -c #(nop) ADD file:45da761e1c56c548be5591e1ec31985006a12a7bff05668c060238638c86adca in / "
	}, {
		"created": "2019-12-26T21:20:11.581830492Z",
		"created_by": "/bin/sh -c #(nop)  CMD [\"sh\"]",
		"empty_layer": true
	}],
	"os": "linux",
	"rootfs": {
		"type": "layers",
		"diff_ids": ["sha256:195be5f8be1df6709dafbba7ce48f2eee785ab7775b88e0c115d8205407265c5"]
	}
}`),
		},
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/library/busybox/blobs/sha256:bdbbaa22dec6b7fe23106d2c1b1f43d9598cd8fc33706cc27c1d938ecd5bffc7",
			Handler: getJSONHandler(`
Content-Type: application/octet-stream
Accept-Ranges: bytes
Age: 551367
Cache-Control: public, max-age=14400
ETag: "4037439e22662d8a453df66a2c025e11"`,
				`binary data`),
		},
		&test.RequestHandlerMapping{
			Method:  "*",
			Pattern: "/v2/library/testupload/blobs",
			Handler: getUploadBlobHandler(),
		},
		&test.RequestHandlerMapping{
			Method:  "*",
			Pattern: "/v2/library/testupload/manifests",
			Handler: getUploadBlobHandler(),
		},
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/",
			Handler: getJSONHandler("", "ok"),
		},
	)

	return server, nil
}

func getJSONHandler(header string, body string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL)
		if buf, e := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: 80}); e == nil {
			fmt.Println("\t", string(buf))
		}
		if body != "" {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		}
		for _, s := range strings.Split(header, "\n") {
			if s = strings.TrimSpace(s); s != "" {
				kv := strings.SplitN(s, ":", 2)
				if len(kv) == 2 {
					w.Header().Add(kv[0], kv[1])
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}
}

func getUploadBlobHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL)
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/blobs/uploads/") {
			header := `
Content-Length: 0
Docker-Distribution-Api-Version: registry/2.0
Docker-Upload-Uuid: 20f99575-cc12-4ad9-a472-2238163d7318
Location: /v2/library/testupload/blobs/uploads/20f99575-cc12-4ad9-a472-2238163d7318
Range: 0-0
Content-Type: text/plain; charset=utf-8
`
			for _, s := range strings.Split(header, "\n") {
				if s = strings.TrimSpace(s); s != "" {
					kv := strings.SplitN(s, ":", 2)
					if len(kv) == 2 {
						w.Header().Add(kv[0], kv[1])
					}
				}
			}
			w.WriteHeader(http.StatusAccepted)
		} else if r.Method == "PATCH" && strings.HasSuffix(r.URL.Path, "/blobs/uploads/20f99575-cc12-4ad9-a472-2238163d7318") {
			if buf, e := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: 80}); e == nil {
				fmt.Println("\t", string(buf))
				w.Header().Add("Range", fmt.Sprintf("0-%d", len(buf)))
			}
			header := `
Content-Length	0
Docker-Distribution-Api-Version	registry/2.0
Docker-Upload-Uuid	20f99575-cc12-4ad9-a472-2238163d7318
Location: /v2/library/testupload/blobs/uploads/20f99575-cc12-4ad9-a472-2238163d7318
Content-Type	text/plain; charset=utf-8`
			for _, s := range strings.Split(header, "\n") {
				if s = strings.TrimSpace(s); s != "" {
					kv := strings.SplitN(s, ":", 2)
					if len(kv) == 2 {
						w.Header().Add(kv[0], kv[1])
					}
				}
			}
			w.WriteHeader(http.StatusAccepted)
		} else if r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/blobs/uploads/20f99575-cc12-4ad9-a472-2238163d7318") {
			header := `
Content-Length	0
Docker-Content-Digest	sha256:bdbbaa22dec6b7fe23106d2c1b1f43d9598cd8fc33706cc27c1d938ecd5bffc7
Docker-Distribution-Api-Version	registry/2.0
Content-Type	text/plain; charset=utf-8`
			for _, s := range strings.Split(header, "\n") {
				if s = strings.TrimSpace(s); s != "" {
					kv := strings.SplitN(s, ":", 2)
					if len(kv) == 2 {
						w.Header().Add(kv[0], kv[1])
					}
				}
			}
			w.WriteHeader(http.StatusCreated)
		} else if r.Method == "HEAD" && strings.HasSuffix(r.URL.Path, "/blobs/sha256:8b190e7dfa3dca5a46811ab688ca0b2f83863a099f86800a88a0eb8b117ac5fe") {
			header := `
Accept-Ranges: bytes
Content-Length: 16
Content-Type: text/plain; charset=utf-8
Docker-Content-Digest: sha256:8b190e7dfa3dca5a46811ab688ca0b2f83863a099f86800a88a0eb8b117ac5fe
Docker-Distribution-Api-Version: registry/2.0`
			for _, s := range strings.Split(header, "\n") {
				if s = strings.TrimSpace(s); s != "" {
					kv := strings.SplitN(s, ":", 2)
					if len(kv) == 2 {
						w.Header().Add(kv[0], kv[1])
					}
				}
			}
			w.WriteHeader(http.StatusOK)
		} else if r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/manifests/latest") {
			header := `
Docker-Content-Digest: sha256:e45c64addef461e06df52ed2818803e35a3e95bc86c494590935a920aee6e4ab
Docker-Distribution-Api-Version: registry/2.0
Location: /v2/library/testupload/manifests/sha256:e45c64addef461e06df52ed2818803e35a3e95bc86c494590935a920aee6e4ab
Content-Length: 0
Content-Type: text/plain; charset=utf-8`
			for _, s := range strings.Split(header, "\n") {
				if s = strings.TrimSpace(s); s != "" {
					kv := strings.SplitN(s, ":", 2)
					if len(kv) == 2 {
						w.Header().Add(kv[0], kv[1])
					}
				}
			}
			w.WriteHeader(http.StatusCreated)
		} else if r.Method == "DELETE" && strings.HasSuffix(r.URL.Path, "/manifests/latest") {
			header := `
Docker-Distribution-Api-Version: registry/2.0
Content-Type: text/plain; charset=utf-8`
			for _, s := range strings.Split(header, "\n") {
				if s = strings.TrimSpace(s); s != "" {
					kv := strings.SplitN(s, ":", 2)
					if len(kv) == 2 {
						w.Header().Add(kv[0], kv[1])
					}
				}
			}
			w.WriteHeader(http.StatusAccepted)
		}
	}
}

func (t *testSuite) SetupTest() {
	mockServer, err := mockRegistry()
	t.Nil(err, "Mock registry failed")

	t.mockRegistry = mockServer
}

func (t *testSuite) TearDownTest() {
	if t.mockRegistry != nil {
		t.mockRegistry.Close()
	}
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
