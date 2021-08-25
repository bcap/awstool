package main

import (
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	opswTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
)

type Result struct {
	EC2      EC2
	ELB      ELB
	Opsworks Opsworks
}

type EC2 struct {
	Reservations []ec2Types.Reservation
	Volumes      []ec2Types.Volume
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

type Opsworks struct {
	Stacks    []opswTypes.Stack
	Layers    []opswTypes.Layer
	Apps      []opswTypes.App
	Instances []opswTypes.Instance
}
