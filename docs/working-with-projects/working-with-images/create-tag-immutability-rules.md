---
title: Tag Immutability Rules
weight: 85
---

By default, users can repeatedly push an image with the same tag to repositories in Harbor. This causes the previous image to effectively be overwritten with each push, in that the tag now points to a different image and the image that previously used the tag now becomes tagless. This is due to the Docker implementation, that does not enforce the mapping between an image tag and the image digest. This can be undesirable in certain cases, because the tag can no longer be trusted to identify the image version. The sha256 digest remains reliable and always points to the same build, but it is not rendered in a human-readable format.

Moreover, the Docker implementation requires that deleting a tag results in the deletion of all other tags that point to the same digest, causing unwanted image deletions.

To prevent this, Harbor allows you to configure tag immutability at the project level, so that images with certain tags cannot be pushed into Harbor if their tags match existing tags. This prevents existing images from being overwritten. Tag immutability guarantees that an immutable tagged image cannot be deleted, and cannot be altered through repushing, retagging, or replication. 

Immutability rules use `OR` logic, so if you set multiple rules and a tag is matched by any of those rules, it is marked as immutable. 

## How Immutable Tags Prevent Tag Deletion

Tags that share a common digest cannot be deleted even if only a single tag is configured as immutable. For example:

1. In a project, set an immutable tag rule that matches the image and tag `hello-world:v1`.
1. In the Docker client, pull `hello-world:v1` and retag it to `hello-world:v2`.
1. Push `hello-world:v2` to the same project.
1. In the Harbor interface, attempt to delete `hello-world:v2`.

In this case, you cannot delete `hello-world:v2` because it shares the sha256 digest with `hello-world:v1`, and `hello-world:v1` is an immutable tag. 

## Create a Tag Immutability Rule

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects**, select a project, and select **Tag Immutability**.

    ![Add an immutability rule](../../../img/tag-immutability.png)

1. Click **Add Rule**.

    - In the **Respositories** row, enter a comma-separated list of repositories to which to either apply or exclude from the rule by selecting either **matching** or **excluding** from the drop-down menu.
    - In the **Tags** row, enter a comma-separated list of tags to which to either apply or exclude from the rule by selecting either **matching** or **excluding** from the drop-down menu.
 
      ![Add an immutability rule](../../../img/add-immutability-rule.png)
1. Click **Add** to save the rule.

    You can add a maximum of 15 immutability rules per project. 

    After you add a rule, any tags that are identified by the rule are marked **Immutable** in the Repositories tab.
1. To modify an existing rule, use the **Action** drop-down menu next to a rule to disable, edit, or delete that rule. 

    ![Immutability rules](../../../img/edit-tag-immutability.png)

## Example

To make all tags for all repositories in the project immutable, set the following options:

- Set **For the respositories** to **matching** and enter `**`.
- Set **Tags** to **matching** and enter `**`.

To allow the tags `rc`, `test`, and `nightly` to be overwritten but make all other tags immutable, set the following options:

- Set **For the respositories** to **matching** and enter `**`.
- Set **Tags** to **excluding** and enter `rc,test,nightly`.
