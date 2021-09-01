```
For Cluster-Provisioning with AccessKey
  1.AWS :
    required items in resource.tf
      -> name  = "NewName"    , cloudProvider Name.
      -> projectname = "defaultproject".
      -> type  = "cluster-provisioning"
      -> providertype = "AWS".
      -> credtype = "accesskey"
      -> accesskey = "xxxxxxxxxxx" ,  // your access key
      -> secretkey = "yyyyyyyyyyy" ,  // your secret key.
      
  2.GCP :
    required items in resource.tf
      -> name  = ""    , cloudProvider Name.
      -> projectname = "defaultproject".
      -> type  = "cluster-provisioning".
      -> providertype = "GCP".
      -> credfile = "<filepath>/gcp-credential.json".
 ```

```
For Cluster-Provisioning with RoleARN
    1. AWS :
      required items in resource.tf
        -> name  = ""    , cloudProvider Name.
        -> projectname = "defaultproject".
        -> type  = "cluster-provisioning"
        -> providertype = "AWS".
        -> credtype = "rolearn".
        -> rolearn = "yyyyyyyyyyyyyyyyy",   // your roleARN value.
        -> externalid = "iiiiiiiiiiiii",    // your externalID value.
  
    2. GCP:
      required items in resource.tf
        -> name  = ""    , cloudProvider Name.
        -> projectname = "defaultproject".
        -> type  = "cluster-provisioning".
        -> providertype = "GCP".
        -> credfile = "<filepath>/gcp-credential.json".
```

```
For Data-Backup with AccessKey
    1. MINIO:
      required items in resource.tf
        -> name  = ""    , cloudProvider Name.
        -> projectname = "defaultproject".
        -> type  = "data-backup".
        -> providertype = "MINIO".
        -> credtype = "accesskey"
        -> accesskey = "xxxxxxxxxxx" ,  // your access key
        -> secretkey = "yyyyyyyyyyy" ,  // your secret key.
  
```
