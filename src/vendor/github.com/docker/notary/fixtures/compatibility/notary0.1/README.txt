$ bin/notary -c cmd/notary/config.json -d /tmp/notary0.1 list docker.com/notary0.1/samplerepo
   NAME                                  DIGEST                                SIZE (BYTES)  
---------------------------------------------------------------------------------------------
  LICENSE   9395bac6fccb26bcb55efb083d1b4b0fe72a1c25f959f056c016120b3bb56a62   11309         


$ bin/notary -c cmd/notary/config.json -d /tmp/notary0.1 status docker.com/notary0.1/samplerepo
Unpublished changes for docker.com/notary0.1/samplerepo:

action    scope     type        path
----------------------------------------------------
create    targets   target      .gitignore
