package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

func deleteSecurityGroups(ctx context.Context, client *ec2.Client, vpcId string, securityGroups []types.SecurityGroup, dryRun bool) (errs error) {
	if dryRun {
		log.Info().Msg("[dryrun]Skipping deletion of SecurityGroup")
		return nil
	}

	for _, securityGroup := range securityGroups {
		if securityGroup.GroupId == nil {
			continue
		}
		if securityGroup.VpcId == nil || *securityGroup.VpcId != vpcId {
			continue
		}

		groupId := *securityGroup.GroupId
		if securityGroupRules, err := listSecurityGroupRules(ctx, client, groupId); err != nil {
			log.Err(err).
				Str("groupId", groupId).
				Msg("listSecurityGroupRules")
		} else {
			log.Info().
				Str("groupId", groupId).
				Strs("securityGroupRuleIds", securityGroupRuleIds(securityGroupRules)).
				Msg("listSecurityGroupRules")
			if len(securityGroupRules) > 0 {
				err := deleteSecurityGroupRules(ctx, client, groupId, securityGroupRules)
				log.Err(err).
					Strs("securityGroupRuleIds", securityGroupRuleIds(securityGroupRules)).
					Msg("deleteSecurityGroupRules")
				if err != nil {
					errs = multierr.Append(errs, err)
					continue
				}
			}
		}

		_, err := client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: securityGroup.GroupId,
		})
		log.Err(err).
			Str("GroupId", groupId).
			Msg("DeleteSecurityGroup")
		errs = multierr.Append(errs, err)
	}
	return
}

func listNonDefaultSecurityGroups(ctx context.Context, client *ec2.Client, vpcId string) ([]types.SecurityGroup, error) {
	input := ec2.DescribeSecurityGroupsInput{
		Filters: ec2VpcFilter(vpcId),
	}
	var securityGroups []types.SecurityGroup
	for {
		output, err := client.DescribeSecurityGroups(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, securityGroup := range output.SecurityGroups {
			if securityGroup.GroupName != nil && *securityGroup.GroupName == "default" {
				continue
			}
			securityGroups = append(securityGroups, securityGroup)
		}
		if output.NextToken == nil {
			return securityGroups, nil
		}
		input.NextToken = output.NextToken
	}
}

func securityGroupIds(securityGroups []types.SecurityGroup) []string {
	securityGroupIds := make([]string, 0, len(securityGroups))
	for _, securityGroup := range securityGroups {
		if securityGroup.GroupId != nil {
			securityGroupIds = append(securityGroupIds, *securityGroup.GroupId)
		}
	}
	return securityGroupIds
}
