package exporter

import (
	"context"
	"fmt"

	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type AwsEC2InstanceIMDs struct {
	up     prometheus.Gauge
	IMDs   *prometheus.Desc
	logger log.Logger
}

type EC2InstanceMetaData struct {
	OwnerId      string
	Name         string
	InstanceID   string
	InstanceType string
}

func (c *AwsEC2InstanceIMDs) ReallyExpensiveAssessmentOfTheSystemState() (
	InstanceIMDs map[string]string, InstanceTags map[string]string,
) {
	// Just example fake data.
	InstanceIMDs = map[string]string{
		"InstanceId":   "12123123",
		"InstanceType": "c3.xlage",
	}
	InstanceTags = map[string]string{
		"Name": "asdf",
	}
	return
}

func (c *AwsEC2InstanceIMDs) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.IMDs
}

func (c *AwsEC2InstanceIMDs) Collect(ch chan<- prometheus.Metric) {
	var up float64
	if instances, err := c.GetInstancesByCache(); err == nil {
		for _, IMD := range instances {
			ch <- prometheus.MustNewConstMetric(
				c.IMDs,
				prometheus.GaugeValue,
				1,
				IMD.InstanceID,
				IMD.Name,
				IMD.OwnerId,
			)
			up = 1
		}
	} else {
		level.Info(c.logger).Log("msg", err.Error())
		up = 0
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName("aws_ec2_imds_exporter", "", "up"), "Was the last scrape of aws ec2 imds successful.", nil, nil),
		prometheus.GaugeValue,
		up,
	)
}

func NewAwsEC2InstanceIMDs(log log.Logger) *AwsEC2InstanceIMDs {
	return &AwsEC2InstanceIMDs{
		IMDs: prometheus.NewDesc(
			prometheus.BuildFQName("aws", "ec2", "info"),
			"aws ec2 info.",
			[]string{"InstanceID", "Name", "OwnerId"},
			nil,
			// prometheus.Labels{"test": "zone"},
		),
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "aws_ec2",
			Name:      "up",
			Help:      "Was the last scrape of haproxy successful.",
		}),
		logger: log,
	}
}

func (c *AwsEC2InstanceIMDs) GetInstancesByCache() ([]EC2InstanceMetaData, error) {
	cache.SetTTL(time.Duration(1 * time.Hour))
	if value, exists := cache.Get("imds"); exists {
		level.Info(c.logger).Log("msg", "use cache")
		return value.([]EC2InstanceMetaData), nil
	}
	IMDs, err := GetInstances()
	if err != nil {
		level.Error(c.logger).Log("msg", err.Error())
		return nil, err
	}
	cache.Set("imds", IMDs)
	level.Info(c.logger).Log("msg", "not use cache")
	return IMDs, nil
}

func GetInstances() ([]EC2InstanceMetaData, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client := ec2.NewFromConfig(cfg)
	input := &ec2.DescribeInstancesInput{}
	result, err := client.DescribeInstances(context.TODO(), input)

	if err != nil {
		return nil, err
	}

	var IMDs []EC2InstanceMetaData
	for _, r := range result.Reservations {
		IMD := EC2InstanceMetaData{}
		IMD.OwnerId = *r.OwnerId
		for _, i := range r.Instances {
			IMD.InstanceID = *i.InstanceId
			for _, t := range i.Tags {
				fmt.Printf("%s: %s \n", *t.Key, *t.Value)
				if *t.Key == "Name" {
					IMD.Name = *t.Value
				}
			}
		}
		IMDs = append(IMDs, IMD)
	}
	return IMDs, nil
}

var cache = ttlcache.NewCache()
