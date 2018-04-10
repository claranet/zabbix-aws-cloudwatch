# Zabbix AWS Cloudwatch

A generic program to get any metric from AWS Cloudwatch depending of provided parameters.
This program handles AWS assume role to get metric from another account.

## Build

Build [zabbix-aws-cloudwatch.go](zabbix-aws-cloudwatch.go):

    $ export GOPATH=$(pwd)/go
    $ cd go/src/zabbix-aws-cloudwatch/
    $ go get
    $ go build

Or simply extract binary from [the bzip2 provided in this repo](https://bitbucket.org/morea/zabbix/downloads/zabbix-aws-cloudwatch.bz2):

    $ bunzip2 zabbix-aws-cloudwatch.bz2

## Installation

Simply copy the `zabbix-aws-cloudwatch` binary to the remote AWS instance from where you want to collect metrics.

## AWS requirements

To work, the instance where this program is run must have the policy `CloudWatchReadOnlyAccess` attached to it from IAM.

## Usage

```
Usage of ./zabbix-aws-cloudwatch:
  -dimensions string
        AWS Cloudwatch dimensions list to filter in Shorthand syntax as for awscli (mandatory)
  -duration string
        AWS Cloudwatch metric duration in string (optional) (default "300s")
  -metric string
        AWS Cloudwatch metric name to collect (mandatory)
  -namespace string
        AWS Cloudwatch namespace to target metric (mandatory)
  -no-data-value string
        Value to returns when there is no data (mandatory)
  -period int
        AWS Cloudwatch metric period in seconds (optional) (default 60)
  -region string
        AWS Cloudwatch region to query (mandatory)
  -role-arn string
        AWS role ARN to assume like arn:aws:iam::myaccountid:role/myrole (optional)
  -stat string
        AWS Cloudwatch metric statistic (mandatory)
```

## Examples

* Get `CreditBurstBalance` from a specific EFS:

```
./zabbix-aws-cloudwatch -region=eu-west-1 -metric=BurstCreditBalance -namespace=AWS/EFS -dimensions='Name=FileSystemId,Value=fs-42424242' -stat=Average -no-data-value=0
1.0440772474677e+13
```

* Get `UnHealthyHostCount` from a specific ELB with multiple dimensions:

```
./zabbix-aws-cloudwatch -region=eu-west-1 -metric=UnHealthyHostCount -namespace=AWS/ELB -dimensions='Name=LoadBalancerName,Value=elb-test Name=AvailabilityZone,Value=eu-west-1a' -stat=Minimum -no-data-value=0
0
```

* Get `HealthyHostCount` from a specific ELB from another account:

```
./zabbix-aws-cloudwatch -region=eu-central-1 -metric=HealthyHostCount -namespace=AWS/ELB -dimensions='Name=LoadBalancerName,Value=elb-test' -role-arn=arn:aws:iam::424213374242:role/iam.ec2.zabbix -no-data-value=0
2
```

## Troubleshooting

* Using assume-role slows down the program compared to no assume-role runtime.
* The program returns 0 whenever either the metric value equals 0 OR the metric is not found (wrong namespace, dimension, metric..)
* Whe you use multiple dimensions, you have to surround the second parameter of the item with double quote (you can do the same for all parameters as best practice)

## References

* [AWS Cloudwatch Concepts](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html)
* [AWS Cloudwatch Metrics and Dimensions Reference](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CW_Support_For_AWS.html)

## License

Copyright (c) 2018 Claranet. Available under the MIT License.