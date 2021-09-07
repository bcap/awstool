package aws

import (
	"awstool/aws/iam"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ebtTypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	esTypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	opswTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	orgTypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWS struct {
	Organization *orgTypes.Organization
	Accounts     map[string]orgTypes.Account
	Regions      map[string]Region
	IAM          iam.IAM
}

func New() AWS {
	return AWS{
		Accounts: map[string]orgTypes.Account{},
		Regions:  map[string]Region{},
		IAM:      iam.New(),
	}
}

type Region struct {
	Region           string
	EC2              EC2
	ELB              ELB
	S3               S3
	Opsworks         Opsworks
	ElasticBeanstalk ElasticBeanstalk
	Elasticsearch    Elasticsearch
}

func NewRegion(region string) Region {
	return Region{
		Region:           region,
		EC2:              NewEC2(),
		ELB:              NewELB(),
		S3:               NewS3(),
		Opsworks:         NewOpsworks(),
		ElasticBeanstalk: NewElasticBeanstalk(),
		Elasticsearch:    NewElasticsearch(),
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

type ElasticsearchDomain struct {
	Status *esTypes.ElasticsearchDomainStatus
	Config *esTypes.ElasticsearchDomainConfig
}

type Elasticsearch struct {
	Domains map[string]*ElasticsearchDomain
}

func NewElasticsearch() Elasticsearch {
	return Elasticsearch{
		Domains: map[string]*ElasticsearchDomain{},
	}
}
