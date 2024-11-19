package lib

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

func ecrClient(ctx context.Context, prefix string) (*awsClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEC2IMDSRegion())
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

func (svc *awsClient) DescribePullThroughCacheRules(ctx context.Context) ([]types.PullThroughCacheRule, error) {
	input := &ecr.DescribePullThroughCacheRulesInput{}
	rules := []types.PullThroughCacheRule{}
	result, err := svc.svc.DescribePullThroughCacheRules(ctx, input)
	if err != nil {
		return rules, fmt.Errorf("list caches: %v", err)
	}

	for _, rule := range result.PullThroughCacheRules {
		if strings.HasPrefix(*rule.EcrRepositoryPrefix, svc.prefix) {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

func ListCaches(ctx context.Context, prefix string, w io.Writer) error {
	svc, err := ecrClient(ctx, prefix)
	if err != nil {
		return err
	}

	rules, err := svc.DescribePullThroughCacheRules(ctx)
	if err != nil {
		return err
	}
	for _, cache := range rules {
		fmt.Fprintf(w, "%s.dkr.ecr.eu-central-1.amazonaws.com/%s -> %s\n", *cache.RegistryId, *cache.EcrRepositoryPrefix, *cache.UpstreamRegistryUrl)
	}

	return nil
}
