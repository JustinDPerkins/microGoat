# Dunder Mifflen Infrastructure Deployment 

<div style="text-align:center">
    <img src="https://raw.githubusercontent.com/JustinDPerkins/microGoat/main/images/microgoat.jpg" alt="Image Alt Text">
</div>



This repository contains an AWS CloudFormation template for deploying Dunder Mifflen's infrastructure. The template provisions resources such as serverless functions, a VPC network, and an ECS cluster for your application. Follow the instructions below to deploy this template.

**This Cloud Application is inherently misconfigured and vulnerable to attack and privilege escalation.**

## 1. Set up for Application Deployment for Github Actions

1. **Fork this repository**.
2. **Launch Stack** using the Launch Stack button below. This will create 5 AWS ECR Image repositories, an OIDC provider for GH, and an IAM Role.
3. **Wait for the stack to complete before proceeding**.
4. Once complete, click the outputs tab and **copy down the two output values**.

Stack Parameters Needed:

- **GitHubOrg**: Name of GitHub organization/user (case sensitive)
- **OIDCAudience**: Audience supplied to configure-aws-credentials. **Just leave the default.**
- **OIDCProviderArn**: Arn for the GitHub OIDC Provider. **Just Ignore, leave it blank.**
- **RepositoryName**: Name of GitHub repository (case sensitive)
- **RepositoryNames**: Comma-separated list of repository names. **Just leave the deafult.**


[![Launch Stack](https://cdn.rawgit.com/buildkite/cloudformation-launch-stack-button-svg/master/launch-stack.svg)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?stackName=DunderRepos&templateURL=https://immersionday-workshops-trendmicro.s3.amazonaws.com/dundermifflen/templates/ecr.template.yaml)

![cft-outputs](/images/Outputs.jpg)

---

## 2. Configure github repo setting.
1. With this repo forked, click the **Settings tab > Secrets and variables > Actions**.
2. Create the following **secrets**.
- **AWS_GH_ROLE_ARN** - Check the Cloudformation OutPuts for the IAM Role ARN for GH to assume.
- **V1_API_KEY** - [How to Generate Vision One API Key](https://docs.trendmicro.com/en-us/documentation/article/trend-vision-one-api-keys)
- **C1_API_KEY** - [How to Generate Cloud One API Key](https://cloudone.trendmicro.com/docs/identity-and-account-management/c1-api-key/#new-api-key)
- **ECR** - Check the Cloudformation OutPuts for the ECR Value. Example "1234567890.dkr.ecr.us-east-1.amazonaws.com"
- **IP_ADDRESS** - Add your own IP. It needs to be in CIDR format. Example: [1.2.3.4/32].
- **LAUNCH_TYPE** - ECS Launch Type. Either '**FARGATE'** or **'EC2'**.

![gh-secrets](/images/gh_secrets.jpg)

---

## 3. Trigger Action Workflow to deploy in AWS
1. Click the **Actions** tab.
2. On the left menu, click **CI/CD Pipeline**.
3. On the Right, click **Run Workflow > Run Workflow**.

----

<details>
  <summary>How To Run Locally</summary>
  
This is a guide on how to build and run the frontend and backend services of my web application using Docker Compose.

## Prerequisites

Before you begin, ensure you have met the following requirements:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)


---

### Run Application Locally

1. Clone this repository to your local machine:

    ```bash
    git clone https://github.com/JustinDPerkins/AirGoatMan.git
    cd AirGoatMan
    ```

2. Return to the project root directory:

    ```bash
    cd deployment
    ```

3. Run the application using Docker Compose:

    ```bash
    docker-compose up --build
    ```

    This command will start both the frontend and backend services and connect them to a shared network.

4. Access the web application in your web browser:

    - Frontend: http://localhost:8080
    - Backend: Your backend API is now accessible via its respective endpoints.

## Stopping the Application

To stop the running containers and remove the associated resources, use the following command:

```bash
docker-compose down
```

---

</details>

## AWS Architecture

![architecture](/images/diagram.png)

---

## Contributors ✨

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/felipecosta09"><img src="https://avatars.githubusercontent.com/u/33869171?v=4" width="100px;" alt=""/><br /><sub><b>Felipe Costa</b></sub></a><br /><a href="https://github.com/JustinDPerkins/microGoat/commits/main/?author=felipecosta09" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/yanmaxsette"><img src="https://avatars.githubusercontent.com/u/31935208?v=4" width="100px;" alt=""/><br /><sub><b>Yan Pinheiro</b></sub></a><br /><a href="https://github.com/JustinDPerkins/microGoat/commits/main/?author=yanmaxsette" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/jmlake569"><img src="https://avatars.githubusercontent.com/u/37003520?v=4" width="100px;" alt=""/><br /><sub><b>Jacob Lake</b></sub></a><br /><a href="https://github.com/JustinDPerkins/microGoat/commits/main/?author=jmlake569" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/JustinDPerkins"><img src="https://avatars.githubusercontent.com/u/60413733?v=4" width="100px;" alt=""/><br /><sub><b>Justin Perkins</b></sub></a><br /><a href="https://github.com/JustinDPerkins/microGoat/commits/main/?author=JustinDPerkins" title="Code">💻</a>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome! 

Thanks also to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):