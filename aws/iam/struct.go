package iam

import (
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type IAM struct {
	Users      []iamTypes.User
	Roles      []Role
	Groups     []iamTypes.Group
	Policies   []iamTypes.Policy
	UserGroups map[string][]iamTypes.Group
	AccessKeys map[string][]iamTypes.AccessKeyMetadata
}

func New() IAM {
	return IAM{
		Users:      []iamTypes.User{},
		Roles:      []Role{},
		Groups:     []iamTypes.Group{},
		Policies:   []iamTypes.Policy{},
		UserGroups: map[string][]iamTypes.Group{},
		AccessKeys: map[string][]iamTypes.AccessKeyMetadata{},
	}
}

type Role struct {
	iamTypes.Role

	AssumeRolePolicyDocument map[string]interface{}
}
