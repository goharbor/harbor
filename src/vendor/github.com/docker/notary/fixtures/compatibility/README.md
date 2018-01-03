This directory contains sample repositories from different versions of Notary client (TUF metadata, trust anchor certificates, and private keys), in order to test backwards compatibility (that newer clients can read old-format repositories).

Notary client makes no guarantees of future-compatibility though (that is, repositories produced by newer clients may not be able to be read by old clients.)

Backwards compatibility has been tested in `client/backwards_compatibility_test.go`

Relevant information for repositories:

- `notary0.1`
	- GUN: `docker.com/notary0.1/samplerepo`
	- key passwords: "randompass"
	- targets:

		```
		   NAME                                  DIGEST                                SIZE (BYTES)
		---------------------------------------------------------------------------------------------
		  LICENSE   9395bac6fccb26bcb55efb083d1b4b0fe72a1c25f959f056c016120b3bb56a62   11309
  		```
  	- It also has a changelist to add a `.gitignore` target, that hasn't been published.

- `notary0.3`
	- GUN: `docker.com/notary0.3/samplerepo`
	- delegations: targets/releases
	- key passwords: "randompass"
	- targets:

		```
		NAME                                  DIGEST                                SIZE (BYTES)         ROLE        
        ----------------------------------------------------------------------------------------------------------------
          LICENSE   9395bac6fccb26bcb55efb083d1b4b0fe72a1c25f959f056c016120b3bb56a62   11309          targets           
          change    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855   0              targets           
          hello     e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855   0              targets/releases
  		```
    - Has a delegation key in the targets/releases role and a corresponding key imported
    - It also has a changelist to add a `MAINTAINERS` target, that hasn't been published to testing publish success.
    - It also has a changelist to add a `Dockerfile` target (an empty file) in the targets/releases role, that hasn't been published to testing publish success with a delegation.
    - unpublished changes:
    
        ```
        Unpublished changes for docker.com/notary0.3/tst:
        
        action    scope     type        path
        ----------------------------------------------------
        create    targets   target      MAINTAINERS
        create    targets/releasestarget      Dockerfile
        ```
    