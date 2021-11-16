package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Ubuntu 20.04 LTS 20211021 hvm:ebs-ssd (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"af-south-1":      {"ami-0ff86122fd4ad7208"},
	"ap-east-1":       {"ami-0a9c1cc3697104990"},
	"ap-northeast-1":  {"ami-036d0684fc96830ca"},
	"ap-northeast-2":  {"ami-0f8b8babb98cc66d0"},
	"ap-northeast-3":  {"ami-0c3904e7363bbc4bc"},
	"ap-south-1":      {"ami-0567e0d2b4b2169ae"},
	"ap-southeast-1":  {"ami-0fed77069cd5a6d6c"},
	"ap-southeast-2":  {"ami-0bf8b986de7e3c7ce"},
	"ca-central-1":    {"ami-0bb84e7329f4fa1f7"},
	"cn-north-1":      {"ami-0741e7b8b4fb0001c"},
	"cn-northwest-1":  {"ami-0883e8062ff31f727"},
	"eu-central-1":     {"ami-0a49b025fffbbdac6"},
	"eu-north-1":      {"ami-0bd9c26722573e69b"},
	"eu-south-1":      {"ami-0f8ce9c417115413d"},
	"eu-west-1":       {"ami-08edbb0e85d6a0a07"},
	"eu-west-2":       {"ami-0fdf70ed5c34c5f52"},
	"eu-west-3":       {"ami-06d79c60d7454e2af"},
	"me-south-1":      {"ami-0b4946d7420c44be4"},
	"sa-east-1":       {"ami-0e66f5495b4efdd0f"},
	"us-east-1":       {"ami-083654bd07b5da81d"},
	"us-east-2":       {"ami-0629230e074c580f2"},
	"us-gov-east-1":   {"ami-0fe6338c47e61cd5d"},
	"us-gov-west-1":   {"ami-087ee83c8de303181"},
	"us-west-1":       {"ami-053ac55bdcfe96e85"},
	"us-west-2":       {"ami-036d46416a34a611c"},
	"custom-endpoint": {""},
}

func awsRegionsList() []string {
	var list []string

	for k := range regionDetails {
		list = append(list, k)
	}

	return list
}

func validateAwsRegion(region string) (string, error) {
	for _, v := range awsRegionsList() {
		if v == region {
			return region, nil
		}
	}

	return "", errors.New("Invalid region specified")
}
