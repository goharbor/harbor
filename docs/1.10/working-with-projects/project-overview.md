---
title: Project Creation
---

A project in Harbor contains all repositories of an application. Images cannot be pushed to Harbor before a project is created. Role-Based Access Control (RBAC) is applied to projects, so that only users with the appropriate roles can perform certain operations. 

There are two types of project in Harbor:  

* **Public**: Any user can pull images from this project. This is a convenient way for you to share repositories with others.
* **Private**: Only users who are members of the project can pull images
