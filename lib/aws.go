package lib

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func ecrClient() (*ecr.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEC2IMDSRegion())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %s ", err)
	}

	svc := ecr.NewFromConfig(cfg)

	return svc, nil
}
