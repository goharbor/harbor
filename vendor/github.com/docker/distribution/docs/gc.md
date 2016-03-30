<!--[metadata]>
+++
title = "Garbage Collection"
description = "High level discussion of garabage collection"
keywords = ["registry, garbage, images, tags, repository, distribution"]
+++
<![end-metadata]-->

# What Garbage Collection Does

Garbage collection is a process that delete blobs to which no manifests refer.
It runs in two phases. First, in the 'mark' phase, the process scans all the 
manifests in the registry. From these manifests, it constructs a set of content 
address digests. This set is the 'mark set' and denotes the set of blobs to *not*
delete. Secondly, in the 'sweep' phase, the process scans all the blobs and if 
a blob's content address digest is not in the mark set, the process will delete 
it.


# How to Run

You can run garbage collection by running

	docker run --rm registry-image-name garbage-collect /etc/docker/registry/config.yml

NOTE: You should ensure that the registry itself is in read-only mode or not running at
all. If you were to upload an image while garbage collection is running, there is the
risk that the image's layers will be mistakenly deleted, leading to a corrupted image.
