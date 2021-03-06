# Zabbix AWS Cloudwatch

A generic program to get any metric from AWS Cloudwatch depending of provided parameters.
This program handles AWS assume role to get metric from another account.

## Installation

Retrieve the last binary version from [releases](https://github.com/claranet/zabbix-aws-cloudwatch/releases) or build it yourself using `go build`.

Then, simply copy the `zabbix-aws-cloudwatch` binary to the remote AWS instance from where you want to collect metrics.

__Note__ there is a [special version](https://github.com/claranet/zabbix-aws-cloudwatch/releases/tag/v1.0.0-go18) `v1.0.0-go18` compiled with go1.8.7 which offers better performance.

## AWS requirements

To work, the instance where this program is run must have the policy `CloudWatchReadOnlyAccess` attached to it from IAM.

## Usage

```
$ ./zabbix-aws-cloudwatch --help
Usage of ./zabbix-aws-cloudwatch:
  -delay string
     AWS Cloudwatch metric delay as string. Ignored if "window" parameter is defined (optional) (default "300s")
  -dimensions string
     AWS Cloudwatch dimensions list to filter in Shorthand syntax as for awscli (mandatory)
  -duration string
     AWS Cloudwatch metric duration as string. Ignored if "window" parameter is defined (optional) (default "60s")
  -metric string
     AWS Cloudwatch metric name to collect (mandatory)
  -namespace string
     AWS Cloudwatch namespace of target metric (mandatory)
  -no-data-value string
     Value to return when there is no data (mandatory)
  -period int
     AWS Cloudwatch metric period in seconds (optional) (default 60)
  -region string
     AWS Cloudwatch region to query (mandatory)
  -role-arn string
     AWS role ARN to assume like arn:aws:iam::myaccountid:role/myrole (optional)
  -stat string
     AWS Cloudwatch metric statistic (mandatory)
  -window string
     AWS Cloudwatch metric window in "duration[:delay]" format like "60s:300s" (optional)
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

* When the "duration" and "period" parameters values chosen involve the cloudwatch API to return multiple points, this program will always return only the last one.
* This program uses a delay of 300s by default to retrieve data from cloudwatch because there is latency before a point is exposed with its final value but it could be decreased according to service refresh time.
* The parameter "window" include both "duration" and "delay" parameters in one. It is useful to allow to bypass the zabbix userparameters limit of 9 parameters.
* If parameter "window" has only one value instead of two ("300s" vs "300s:900s") default delay value (300s) will be used for retro compatibility.
* Using assume-role slows down the program compared to no assume-role runtime.
* The program returns 0 whenever either the metric value equals 0 OR the metric is not found (wrong namespace, dimension, metric..)
* When you use multiple dimensions, you have to surround the second parameter of the item with double quote (you can do the same for all parameters as best practice)
* You can test new items creation with the provided [userparameter](../../../zabbix_agentd.d/aws.conf) using the following commands :

```
zabbix_agentd -t 'aws.cloudwatch.metric["AWS/ELB","Name=LoadBalancerName,Value=elb-test Name=AvailabilityZone,Value=eu-west-1a","HealthyHostCount","Minimum","300s:300s","300","eu-west-1",0,]'
zabbix_get -s 127.0.0.1 -k 'aws.cloudwatch.metric["AWS/ApplicationELB","Name=LoadBalancer,Value=app/alb-test/371554445f0edb42","RequestCount","Sum","60s:120s","60","eu-west-1",0]'
```

## References

* [AWS Cloudwatch Concepts](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html)
* [AWS Cloudwatch Metrics and Dimensions Reference](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CW_Support_For_AWS.html)

## License

Copyright (c) 2018 Claranet. Available under the MIT License.
