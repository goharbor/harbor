GO_VERSION=1.5.1
GO_SRC=/tmp/go/src
PATCH_DIR=$(CURDIR)/patches

.PHONY: clean download_go_src run_patch update

define apply_patch
mkdir -p $(CURDIR)/$(2);
cp -r ${GO_SRC}/$(1)/* $(CURDIR)/$(2)/;
git apply $(PATCH_DIR)/*;
endef

download_go_src:
	curl -sSL https://golang.org/dl/go${GO_VERSION}.src.tar.gz | tar -v -C /tmp -xz

run_patch: download_go_src
	@rm -rf canonical/json
	$(call apply_patch,encoding/json,canonical/json,json)

update: run_patch clean

clean:
	rm -rf ${GO_SRC}
