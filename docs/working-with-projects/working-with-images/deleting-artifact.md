---
title: Deleting Artifacts
weight: 75
---

Harbor v2.0 now supports OCI images and OCI image indexes (https://github.com/opencontainers/image-spec/blob/master/image-index.md). An OCI image index (or OCI index) is a higher level manifest which points to a list of image manifests, ideal for one or more platforms.  Both the index itself and the images referenced within are called artifacts in Harbor. An OCI index can hold another OCI index and so on and so forth. For any artifact referenced by an OCI index, the referenced artifact is known as the child artifact and the OCI index referencing the artifact is known as the parent artifact. The child artifact belongs to the parent artifact or is a part of the parent artifact.  

An example of an OCI image index 

```
{
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7143,
      "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7682,
      "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    }
  ],
  "annotations": {
    "com.example.key1": "value1",
    "com.example.key2": "value2"
  }
}
```

**Deleting an Artifact**

When an artifact is not referenced by any OCI index, you can delete the artifact freely which also deletes its manifest and all associated tags. 

When an artifact is referenced by an OCI index, you cannot delete it.  To delete a referenced artifact, you must first delete all OCI indexes referencing the artifact. Remember that an artifact can be referenced by multiple parent artifacts pushed onto Harbor by different users.  So when you delete an OCI index holding 9 child artifacts that are not referenced by any other index and 1 child artifact referenced by another index, only 9 out of 10 child artifacts are deleted.

To delete an artifact in the Harbor interface, click on the artifact and select **Delete**, and then confirm.  

![delete image1](../../../img/deleteimage1.png)

![delete image2](../../../img/deleteimage2.png)
