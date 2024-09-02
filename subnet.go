package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

func deleteSubnets(ctx context.Context, client *ec2.Client, vpcId string, subnets []types.Subnet, dryRun bool) (errs error) {
	if dryRun {
		log.Info().Msg("[dryrun]Skipping deletion of Subnet")
		return nil
	}

	for _, subnet := range subnets {
		if subnet.SubnetId == nil {
			continue
		}
		if subnet.VpcId == nil || *subnet.VpcId != vpcId {
			continue
		}

		_, err := client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		log.Err(err).
			Str("SubnetId", *subnet.SubnetId).
			Msg("DeleteSubnet")
		errs = multierr.Append(errs, err)
	}
	return
}

func listSubnets(ctx context.Context, client *ec2.Client, vpcId string) ([]types.Subnet, error) {
	input := ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	}
	var subnets []types.Subnet
	for {
		output, err := client.DescribeSubnets(ctx, &input)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, output.Subnets...)
		if output.NextToken == nil {
			return subnets, nil
		}
		input.NextToken = output.NextToken
	}
}

func subnetIds(subnets []types.Subnet) []string {
	subnetIds := make([]string, 0, len(subnets))
	for _, subnet := range subnets {
		if subnet.SubnetId != nil {
			subnetIds = append(subnetIds, *subnet.SubnetId)
		}
	}
	return subnetIds
}
