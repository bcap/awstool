package dump

import (
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ebtTypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	opswTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWS struct {
	Regions map[string]AWSRegion
	IAM     IAM
}

func NewAWS() AWS {
	return AWS{
		Regions: map[string]AWSRegion{},
		IAM:     NewIAM(),
	}
}

type AWSRegion struct {
	Region           string
	EC2              EC2
	ELB              ELB
	S3               S3
	Opsworks         Opsworks
	ElasticBeanstalk ElasticBeanstalk
}

func NewAWSRegion(region string) AWSRegion {
	return AWSRegion{
		Region:           region,
		EC2:              NewEC2(),
		ELB:              NewELB(),
		S3:               NewS3(),
		Opsworks:         NewOpsworks(),
		ElasticBeanstalk: NewElasticBeanstalk(),
	}
}

type IAM struct {
	Users      []iamTypes.User
	Roles      []iamTypes.Role
	Groups     []iamTypes.Group
	Policies   []iamTypes.Policy
	UserGroups map[string][]iamTypes.Group
	AccessKeys map[string][]iamTypes.AccessKeyMetadata
}

func NewIAM() IAM {
	return IAM{
		Users:      []iamTypes.User{},
		Roles:      []iamTypes.Role{},
		Groups:     []iamTypes.Group{},
		Policies:   []iamTypes.Policy{},
		UserGroups: map[string][]iamTypes.Group{},
		AccessKeys: map[string][]iamTypes.AccessKeyMetadata{},
	}
}

type EC2 struct {
	Reservations []ec2Types.Reservation
	Volumes      []ec2Types.Volume
}

func NewEC2() EC2 {
	return EC2{
		Reservations: []ec2Types.Reservation{},
		Volumes:      []ec2Types.Volume{},
	}
}

type ELB struct {
	V1 ELBv1
	V2 ELBv2
}

type ELBv1 struct {
	LoadBalancers []elbTypes.LoadBalancerDescription
}

type ELBv2 struct {
	LoadBalancers []elbv2Types.LoadBalancer
}

func NewELB() ELB {
	return ELB{
		V1: ELBv1{
			LoadBalancers: []elbTypes.LoadBalancerDescription{},
		},
		V2: ELBv2{
			LoadBalancers: []elbv2Types.LoadBalancer{},
		},
	}
}

type S3 struct {
	Buckets    []s3Types.Bucket
	BucketTags map[string][]s3Types.Tag
}

func NewS3() S3 {
	return S3{
		Buckets:    []s3Types.Bucket{},
		BucketTags: map[string][]s3Types.Tag{},
	}
}

type Opsworks struct {
	Stacks    []opswTypes.Stack
	Layers    []opswTypes.Layer
	Apps      []opswTypes.App
	Instances []opswTypes.Instance
}

func NewOpsworks() Opsworks {
	return Opsworks{
		Stacks:    []opswTypes.Stack{},
		Layers:    []opswTypes.Layer{},
		Apps:      []opswTypes.App{},
		Instances: []opswTypes.Instance{},
	}
}

type ElasticBeanstalk struct {
	Applications []ebtTypes.ApplicationDescription
	Environments []ebtTypes.EnvironmentDescription
}

func NewElasticBeanstalk() ElasticBeanstalk {
	return ElasticBeanstalk{
		Applications: []ebtTypes.ApplicationDescription{},
		Environments: []ebtTypes.EnvironmentDescription{},
	}
}
