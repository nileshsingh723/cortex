/*
Copyright 2019 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusterconfig

import (
	"strings"

	"github.com/cortexlabs/cortex/pkg/consts"
	"github.com/cortexlabs/cortex/pkg/lib/aws"
	cr "github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
	"github.com/cortexlabs/cortex/pkg/lib/hash"
	"github.com/cortexlabs/cortex/pkg/lib/pointer"
	"github.com/cortexlabs/cortex/pkg/lib/prompt"
	"github.com/cortexlabs/cortex/pkg/lib/table"
)

type ClusterConfig struct {
	InstanceType           *string `json:"instance_type" yaml:"instance_type"`
	MinInstances           *int64  `json:"min_instances" yaml:"min_instances"`
	MaxInstances           *int64  `json:"max_instances" yaml:"max_instances"`
	ClusterName            string  `json:"cluster_name" yaml:"cluster_name"`
	Region                 string  `json:"region" yaml:"region"`
	Bucket                 string  `json:"bucket" yaml:"bucket"`
	LogGroup               string  `json:"log_group" yaml:"log_group"`
	InstanceVolumeSize     int64   `json:"instance_volume_size" yaml:"instance_volume_size"`
	Telemetry              bool    `json:"telemetry" yaml:"telemetry"`
	ImagePredictorServe    string  `json:"image_predictor_serve" yaml:"image_predictor_serve"`
	ImagePredictorServeGPU string  `json:"image_predictor_serve_gpu" yaml:"image_predictor_serve_gpu"`
	ImageTFServe           string  `json:"image_tf_serve" yaml:"image_tf_serve"`
	ImageTFServeGPU        string  `json:"image_tf_serve_gpu" yaml:"image_tf_serve_gpu"`
	ImageTFAPI             string  `json:"image_tf_api" yaml:"image_tf_api"`
	ImageONNXServe         string  `json:"image_onnx_serve" yaml:"image_onnx_serve"`
	ImageONNXServeGPU      string  `json:"image_onnx_serve_gpu" yaml:"image_onnx_serve_gpu"`
	ImageOperator          string  `json:"image_operator" yaml:"image_operator"`
	ImageManager           string  `json:"image_manager" yaml:"image_manager"`
	ImageDownloader        string  `json:"image_downloader" yaml:"image_downloader"`
	ImageClusterAutoscaler string  `json:"image_cluster_autoscaler" yaml:"image_cluster_autoscaler"`
	ImageMetricsServer     string  `json:"image_metrics_server" yaml:"image_metrics_server"`
	ImageNvidia            string  `json:"image_nvidia" yaml:"image_nvidia"`
	ImageFluentd           string  `json:"image_fluentd" yaml:"image_fluentd"`
	ImageStatsd            string  `json:"image_statsd" yaml:"image_statsd"`
	ImageIstioProxy        string  `json:"image_istio_proxy" yaml:"image_istio_proxy"`
	ImageIstioPilot        string  `json:"image_istio_pilot" yaml:"image_istio_pilot"`
	ImageIstioCitadel      string  `json:"image_istio_citadel" yaml:"image_istio_citadel"`
	ImageIstioGalley       string  `json:"image_istio_galley" yaml:"image_istio_galley"`
}

type InternalClusterConfig struct {
	ClusterConfig
	ID                string `json:"id"`
	APIVersion        string `json:"api_version"`
	OperatorInCluster bool   `json:"operator_in_cluster"`
}

var Validation = &cr.StructValidation{
	StructFieldValidations: []*cr.StructFieldValidation{
		{
			StructField: "InstanceType",
			StringPtrValidation: &cr.StringPtrValidation{
				Validator: validateInstanceType,
			},
		},
		{
			StructField: "MinInstances",
			Int64PtrValidation: &cr.Int64PtrValidation{
				GreaterThan: pointer.Int64(0),
			},
		},
		{
			StructField: "MaxInstances",
			Int64PtrValidation: &cr.Int64PtrValidation{
				GreaterThan: pointer.Int64(0),
			},
		},
		{
			StructField: "ClusterName",
			StringValidation: &cr.StringValidation{
				Default: "cortex",
			},
		},
		{
			StructField: "Region",
			StringValidation: &cr.StringValidation{
				Default: "us-west-2",
			},
		},
		{
			StructField: "Bucket",
			StringValidation: &cr.StringValidation{
				Default:    "",
				AllowEmpty: true,
			},
		},
		{
			StructField: "LogGroup",
			StringValidation: &cr.StringValidation{
				Default: "cortex",
			},
		},
		{
			StructField: "InstanceVolumeSize",
			Int64Validation: &cr.Int64Validation{
				Default:              50,
				GreaterThanOrEqualTo: pointer.Int64(20), // large enough to fit docker images and any other overhead
				LessThanOrEqualTo:    pointer.Int64(16384),
			},
		},
		{
			StructField: "Telemetry",
			BoolValidation: &cr.BoolValidation{
				Default: true,
			},
		},
		{
			StructField: "ImagePredictorServe",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/predictor-serve:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImagePredictorServeGPU",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/predictor-serve-gpu:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageTFServe",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/tf-serve:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageTFServeGPU",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/tf-serve-gpu:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageTFAPI",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/tf-api:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageONNXServe",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/onnx-serve:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageONNXServeGPU",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/onnx-serve-gpu:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageOperator",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/operator:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageManager",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/manager:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageDownloader",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/downloader:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageClusterAutoscaler",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/cluster-autoscaler:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageMetricsServer",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/metrics-server:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageNvidia",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/nvidia:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageFluentd",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/fluentd:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageStatsd",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/statsd:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageIstioProxy",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/istio-proxy:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageIstioPilot",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/istio-pilot:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageIstioCitadel",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/istio-citadel:" + consts.CortexVersion,
			},
		},
		{
			StructField: "ImageIstioGalley",
			StringValidation: &cr.StringValidation{
				Default: "cortexlabs/istio-galley:" + consts.CortexVersion,
			},
		},
		// Extra keys that exist in the cluster config file
		{
			Key: "aws_access_key_id",
			Nil: true,
		},
		{
			Key: "aws_secret_access_key",
			Nil: true,
		},
		{
			Key: "cortex_aws_access_key_id",
			Nil: true,
		},
		{
			Key: "cortex_aws_secret_access_key",
			Nil: true,
		},
	},
}

func PromptValidation(skipPopulatedFields bool, promptInstanceType bool, defaults *ClusterConfig) *cr.PromptValidation {
	if defaults == nil {
		defaults = &ClusterConfig{}
	}
	if defaults.InstanceType == nil {
		defaults.InstanceType = pointer.String("m5.large")
	}
	if defaults.MinInstances == nil {
		defaults.MinInstances = pointer.Int64(2)
	}
	if defaults.MaxInstances == nil {
		defaults.MaxInstances = pointer.Int64(5)
	}

	var promptItemValidations []*cr.PromptItemValidation

	if promptInstanceType {
		promptItemValidations = append(promptItemValidations,
			&cr.PromptItemValidation{
				StructField: "InstanceType",
				PromptOpts: &prompt.Options{
					Prompt: "AWS instance type",
				},
				StringPtrValidation: &cr.StringPtrValidation{
					Required:  true,
					Default:   defaults.InstanceType,
					Validator: validateInstanceType,
				},
			},
		)
	}

	promptItemValidations = append(promptItemValidations,
		&cr.PromptItemValidation{
			StructField: "MinInstances",
			PromptOpts: &prompt.Options{
				Prompt: "Min instances",
			},
			Int64PtrValidation: &cr.Int64PtrValidation{
				Required:    true,
				Default:     defaults.MinInstances,
				GreaterThan: pointer.Int64(0),
			},
		},
		&cr.PromptItemValidation{
			StructField: "MaxInstances",
			PromptOpts: &prompt.Options{
				Prompt: "Max instances",
			},
			Int64PtrValidation: &cr.Int64PtrValidation{
				Required:    true,
				Default:     defaults.MaxInstances,
				GreaterThan: pointer.Int64(0),
			},
		},
	)

	return &cr.PromptValidation{
		SkipPopulatedFields:   skipPopulatedFields,
		PromptItemValidations: promptItemValidations,
	}
}

func validateInstanceType(instanceType string) (string, error) {
	if strings.HasSuffix(instanceType, "nano") ||
		strings.HasSuffix(instanceType, "micro") ||
		strings.HasSuffix(instanceType, "small") {
		return "", ErrorInstanceTypeTooSmall()
	}

	return instanceType, nil
}

// This does not set defaults for fields that are prompted from the user
func SetFileDefaults(clusterConfig *ClusterConfig) error {
	var emtpyMap interface{} = map[interface{}]interface{}{}
	errs := cr.Struct(clusterConfig, emtpyMap, Validation)
	if errors.HasErrors(errs) {
		return errors.FirstError(errs...)
	}
	return nil
}

// This does not set defaults for fields that are prompted from the user
func GetFileDefaults() (*ClusterConfig, error) {
	clusterConfig := &ClusterConfig{}
	err := SetFileDefaults(clusterConfig)
	if err != nil {
		return nil, err
	}

	return clusterConfig, nil
}

func (cc *ClusterConfig) SetBucket(awsAccessKeyID string, awsSecretAccessKey string) error {
	if cc.Bucket != "" {
		return nil
	}

	awsAccountID, validCreds, err := aws.AccountID(awsAccessKeyID, awsSecretAccessKey, cc.Region)
	if err != nil {
		return err
	}
	if !validCreds {
		return ErrorInvalidAWSCredentials()
	}

	cc.Bucket = "cortex-" + hash.String(awsAccountID)[:10]
	return nil
}

func (cc *InternalClusterConfig) String() string {
	var items []table.KV

	items = append(items, table.KV{K: "cluster version", V: cc.APIVersion})
	items = append(items, table.KV{K: "instance type", V: *cc.InstanceType})
	items = append(items, table.KV{K: "min instances", V: *cc.MinInstances})
	items = append(items, table.KV{K: "max instances", V: *cc.MaxInstances})
	items = append(items, table.KV{K: "cluster name", V: cc.ClusterName})
	items = append(items, table.KV{K: "region", V: cc.Region})
	items = append(items, table.KV{K: "bucket", V: cc.Bucket})
	items = append(items, table.KV{K: "log group", V: cc.LogGroup})
	items = append(items, table.KV{K: "instance volume size", V: cc.InstanceVolumeSize})
	items = append(items, table.KV{K: "telemetry", V: cc.Telemetry})
	items = append(items, table.KV{K: "image_predictor_serve", V: cc.ImagePredictorServe})
	items = append(items, table.KV{K: "image_predictor_serve_gpu", V: cc.ImagePredictorServeGPU})
	items = append(items, table.KV{K: "image_tf_serve", V: cc.ImageTFServe})
	items = append(items, table.KV{K: "image_tf_serve_gpu", V: cc.ImageTFServeGPU})
	items = append(items, table.KV{K: "image_tf_api", V: cc.ImageTFAPI})
	items = append(items, table.KV{K: "image_onnx_serve", V: cc.ImageONNXServe})
	items = append(items, table.KV{K: "image_onnx_serve_gpu", V: cc.ImageONNXServeGPU})
	items = append(items, table.KV{K: "image_operator", V: cc.ImageOperator})
	items = append(items, table.KV{K: "image_manager", V: cc.ImageManager})
	items = append(items, table.KV{K: "image_downloader", V: cc.ImageDownloader})
	items = append(items, table.KV{K: "image_cluster_autoscaler", V: cc.ImageClusterAutoscaler})
	items = append(items, table.KV{K: "image_metrics_server", V: cc.ImageMetricsServer})
	items = append(items, table.KV{K: "image_nvidia", V: cc.ImageNvidia})
	items = append(items, table.KV{K: "image_fluentd", V: cc.ImageFluentd})
	items = append(items, table.KV{K: "image_statsd", V: cc.ImageStatsd})
	items = append(items, table.KV{K: "image_istio_proxy", V: cc.ImageIstioProxy})
	items = append(items, table.KV{K: "image_istio_pilot", V: cc.ImageIstioPilot})
	items = append(items, table.KV{K: "image_istio_citadel", V: cc.ImageIstioCitadel})
	items = append(items, table.KV{K: "image_istio_galley", V: cc.ImageIstioGalley})

	return table.AlignKeyValue(items, ":", 1)
}
