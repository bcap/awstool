package organizations

import (
	"awstool/common"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgTypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	log "github.com/sirupsen/logrus"
)

func FetchOrganization(ctx context.Context, cfg aws.Config) (orgTypes.Organization, error) {
	log.Debug("Fetching organization")
	client := organizations.NewFromConfig(cfg)
	result, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		return orgTypes.Organization{}, err
	}
	log.Info("Fetched organization")
	return *result.Organization, nil
}

func FetchAllAccounts(ctx context.Context, cfg aws.Config) ([]orgTypes.Account, error) {
	log.Debug("Fetching all AWS accounts")
	client := organizations.NewFromConfig(cfg)
	accounts := []orgTypes.Account{}
	load := func(nextToken *string) (*string, error) {
		result, err := client.ListAccounts(ctx, &organizations.ListAccountsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, result.Accounts...)
		return result.NextToken, nil
	}
	common.FetchAll("accounts", load)
	log.Infof("Fetched %d AWS accounts", len(accounts))
	return accounts, nil
}
