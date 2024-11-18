package lib

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func ecrClient(prefix string) (*awsClient, error) {
	// TODO use app ctx
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEC2IMDSRegion())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %s ", err)
	}

	svc := ecr.NewFromConfig(cfg)

	return &awsClient{svc, prefix}, nil
}

type awsClient struct {
	svc    *ecr.Client
	prefix string
}

func (svc *awsClient) GetAuthorizationToken(ctx context.Context) (*ecr.GetAuthorizationTokenOutput, error) {
	return svc.svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
}

func (svc *awsClient) DescribePullThroughCacheRules(ctx context.Context) (*ecr.DescribePullThroughCacheRulesOutput, error) {
	input := &ecr.DescribePullThroughCacheRulesInput{EcrRepositoryPrefixes: []string{svc.prefix}}
	result, err := svc.svc.DescribePullThroughCacheRules(ctx, input)

	return result, err
}
