---
title: Retagging Images
weight: 75
---

You can retag images in Harbor. Users with sufficient privileges can retag images to different repositories and projects. For example, you can retag images as follows:

- `release/app:stg`  -->  `release/app:prd`
- `develop/app:v1.0` --> `release/app:v1.0`

To retag an image, you must have read permission (guest role or above) in the source project and write permission (developer role or above) in the target project.

In the Harbor interface, select the image to retag, and click `Retag`.

![retag image](../../img/retag-image.png)

In the Retag windown, enter the project name, repository name, the new tag name, and click **Confirm**. 
