---
title: Detagging Artifacts
weight: 75
---

Harbor v2.0 supports OCI images and OCI image indexes (https://github.com/opencontainers/image-spec/blob/master/image-index.md). An OCI image index (or OCI index) is a higher level manifest which points to a list of image manifests, ideal for one or more platforms.  Both the index itself and the images referenced within are called artifacts in Harbor. An OCI index can hold another OCI index and so on.  For any artifact referenced by an OCI index, the referenced artifact is known as the child artifact and the OCI index referencing the artifact is known as the parent artifact. The child artifact belongs to the parent artifact or is a part of the parent artifact.  

You can delete an existing tag from an artifact without deleting the artifact digest and all other existing tags. For an OCI index, you can delete tags from the parent artifact as well as from the referenced artifacts within. Tags removed from the parent artifact are not automatically removed from its child artifacts.

In the Harbor interface, select an artifact to see its current set of tags, then select the tag you want to delete and click **REMOVE TAG**, and then confirm.

![delete tag](../../../img/deletetag1.png)

You can remove all tags from an artifact without deleting the artifact manifest itself.  The artifact is still visible on the web console with nothing listed under **Tags**.
