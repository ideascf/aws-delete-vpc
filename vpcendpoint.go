package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

func deleteVpcEndpoints(ctx context.Context, client *ec2.Client, vpcId string, vpcPeeringConnections []types.VpcEndpoint, dryRun bool) (errs error) {
	if dryRun {
		log.Info().Msg("[dryrun]Skipping deletion of VpcPeering")
		return nil
	}

	for _, vpcPeeringConnection := range vpcPeeringConnections {
		if vpcPeeringConnection.VpcEndpointId == nil {
			continue
		}

		_, err := client.DeleteVpcEndpoints(ctx, &ec2.DeleteVpcEndpointsInput{
			VpcEndpointIds: []string{*vpcPeeringConnection.VpcEndpointId},
		})
		log.Err(err).
			Str("VpcEndpointId", *vpcPeeringConnection.VpcEndpointId).
			Msg("DeleteVpcEndpoint")
		errs = multierr.Append(errs, err)
	}
	return
}

func listVpcEndpoints(ctx context.Context, client *ec2.Client, vpcId string) ([]types.VpcEndpoint, error) {
	var vpcEndpoints []types.VpcEndpoint

	input := ec2.DescribeVpcEndpointsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	}
	for {
		output, err := client.DescribeVpcEndpoints(ctx, &input)
		if err != nil {
			return nil, err
		}
		vpcEndpoints = append(vpcEndpoints, output.VpcEndpoints...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}
	return vpcEndpoints, nil
}

func vpcEndpointIds(vpcEndpoints []types.VpcEndpoint) []string {
	vpcEndpointIds := make([]string, 0, len(vpcEndpoints))
	for _, vpcEndpoint := range vpcEndpoints {
		if vpcEndpoint.VpcEndpointId != nil {
			var name string
			for _, tag := range vpcEndpoint.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
					break
				}
			}
			vpcEndpointIds = append(vpcEndpointIds, fmt.Sprintf("%s:%s", *vpcEndpoint.VpcEndpointId, name))
		}
	}
	return vpcEndpointIds
}
