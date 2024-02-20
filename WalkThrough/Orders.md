# Boring Paper Company: Orders Exploit Walkthrough

This webpage is backed by a Container running in ECS as well as serving content to and from S3. When a upload request is made the API forwards the object to S3. The second feature is select content is served to the user from S3 in the case the user needs to retrieve an object. 

Certain applications types may request or server objects to its clients.

However, not all objects may be safe or how can we be sure they are safe for distribution.

---

### Discovery:

Right-click on the web page, and select "Inspect" or "Inspect Element" from the context menu. This will open the browser's developer tools

- Click on the Network tab, in the inspector tools window. 
- Now with the browser developer inspection view open, refresh the page to capture page load requests.
- Expand the 200 status GET method event.
- Discover the backend S3 Endpoint URL where objects are being stored for this application.


The format should be: https://boringpaperco-warehousestack-SomeRandomString.s3.amazonaws.com

![lambda](/images/orders/s3.jpg)

---

### S3 Resource Misconfiguration: Bucket's List access permission set to "Everyone"

You now know that we have a bucket. From the Loadbalancer we see `us-east-1`, so you can attempt to browse the bucket by using the aws cli by running:

```
aws s3 ls  s3://discovered-bucket-name-here/ --no-sign-request --region us-east-1
```

---

### S3 Capture Secret Object: Data Exfiltration

```
aws s3 cp s3://discovered-bucket-name-here/secret.txt --no-sign-request --region us-east-2 secret.html
```

---

### Persistance: S3 Misconfigured bucket PutObject allowed for 'Everyone'

```
aws s3 cp <path-to-file>/somefile.pdf --no-sign-request --region us-east-2 s3://bucketnamehere/downloads/BORING_PAPER_REQUEST_FORM.pdf
```


---


### Explanation 

Here is the underlying Infrastructure code running for S3.

```
WareHouseOrders:
    Type: AWS::S3::Bucket
    Properties:
      PublicAccessBlockConfiguration:
        BlockPublicAcls: false
        BlockPublicPolicy: false
        IgnorePublicAcls: false
        RestrictPublicBuckets: false

  WareHousePolicy:
    Type: 'AWS::S3::BucketPolicy'
    Properties:
      Bucket:
        Ref: 'WareHouseOrders'
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal: '*'
            Action: 's3:*'
            Resource:
              Fn::Join:
                - ''
                - - 'arn:aws:s3:::'
                  - Ref: 'WareHouseOrders'
                  - '/*' 
```

---

Let's break down the issues:

**Public Access Configuration:**
    Issue: All four parameters (BlockPublicAcls, BlockPublicPolicy, IgnorePublicAcls, RestrictPublicBuckets) in PublicAccessBlockConfiguration are set to false. This means that public access is not blocked, and anyone with the bucket name can potentially access it.
    Recommendation: Set these parameters to true to block public access and enhance security.

**Bucket Policy Principal:**
    Issue: The bucket policy allows any principal (```*```) to perform any S3 action (s3:```*```) on the specified resources. This includes the ability for random users to place objects in the bucket.
    Recommendation: Restrict access to specific AWS accounts, IAM users, or roles by specifying them in the Principal field. Allowing random users (```*```) to perform any action is a security risk.

---

