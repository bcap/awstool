package awstool

// This file is used to speed up docker builds
// By building only this file before the app code, we force
// building dependencies and caching them
// This file should not be included in the final generated image

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var _ aws.HTTPClient
var _ cobra.Command
var _ config.Config
var _ ec2.Client
var _ elasticbeanstalk.Client
var _ elasticloadbalancing.Client
var _ elasticloadbalancingv2.Client
var _ elasticsearchservice.Client
var _ iam.Client
var _ logrus.Level
var _ opsworks.Client
var _ organizations.Client
var _ s3.Client
var _ semaphore.Weighted
var _ spew.ConfigState
