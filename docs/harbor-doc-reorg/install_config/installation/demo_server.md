# Test Harbor with the Demo Server

The Harbor team has made available a demo Harbor instance that you can use to experiment with Harbor and test its functionalities.

When using the demo server, please take note of the conditions of use.

## Conditions of Use of the Demo Server ##

 - The demo server is reserved for experimental use only, to allow you to test Harbor functionality. 
 - Do not upload sensitive images to the demo server. 
 - The demo server is not a production environment. The Harbor team is not responsible for any loss of data, functionality, or service that might result from its use.
 - The demo server is cleaned and reset every two days.
 - The demo server only allows you to test user functionalities. You cannot test administrator functionalities. To test administrator functionalities and advanced features, set up a Harbor instance.
 - Do not push images >100MB to the demo server, as it has limited storage capacity.

If you encounter any problems while using the demo server, open an [issue on Github](https://github.com/goharbor/harbor/issues) or contact the Harbor team on [Slack](https://github.com/goharbor/harbor#community).

## Access the Demo Server ##

1. Go to  [https://demo.goharbor.io](https://demo.goharbor.io).
1. Click **Sign up for an account**.
1. Create a user account by providing a username, your email address, your name, and a password.
1. Log in to the Harbor interface using the account you created.
1. Explore the default project, `library` and create your own project.

   For information about how to create a project, see [Managing Projects](../../working_with_projects/managing_projects.md).
1. Open a Docker client and log in to Harbor.

   ```
   docker login demo.goharbor.io
   ```
1. Build an image, tag it, and push it to a project in Harbor.

   ```
   docker push demo.goharbor.io/your-project/test-container
   ```   
1. In the Harbor interface, go to the project and select the **Repositories** tab to view the image repository in the Harbor project.