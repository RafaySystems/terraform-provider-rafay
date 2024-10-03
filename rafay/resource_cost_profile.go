package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/costpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCostProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCostProfileCreate,
		ReadContext:   resourceCostProfileRead,
		UpdateContext: resourceCostProfileUpdate,
		DeleteContext: resourceCostProfileDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.CostProfileSchema.Schema,
	}
}

func resourceCostProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("CostProfile create starts")
	diags := resourceCostProfileUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("costProfile create got error, perform cleanup")
		mp, err := expandCostProfile(d)
		if err != nil {
			log.Printf("costProfile expandCostProfile error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.CostV3().CostProfile().Delete(ctx, options.DeleteOptions{
			Name:    mp.Metadata.Name,
			Project: mp.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceCostProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("costProfile update starts")
	return resourceCostProfileUpsert(ctx, d, m)
}

func resourceCostProfileUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("costProfile upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	costProfile, err := expandCostProfile(d)
	if err != nil {
		log.Printf("costProfile expandCostProfile error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CostV3().CostProfile().Apply(ctx, costProfile, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", costProfile)
		log.Println("costProfile apply costProfile:", n1)
		log.Printf("costProfile apply error")
		return diag.FromErr(err)
	}

	d.SetId(costProfile.Metadata.Name)
	return diags

}

func resourceCostProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceCostProfileRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	tfCostProfileState, err := expandCostProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	mp, err := client.CostV3().CostProfile().Get(ctx, options.GetOptions{
		//Name:    tfCostProfileState.Metadata.Name,
		Name:    meta.Name,
		Project: tfCostProfileState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenCostProfile(d, mp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceCostProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	mp, err := expandCostProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CostV3().CostProfile().Delete(ctx, options.DeleteOptions{
		Name:    mp.Metadata.Name,
		Project: mp.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandCostProfile(in *schema.ResourceData) (*costpb.CostProfile, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand costProfile empty input")
	}
	obj := &costpb.CostProfile{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandCostProfileSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandCostProfileSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "cost.k8smgmt.io/v3"
	obj.Kind = "CostProfile"
	return obj, nil
}

func expandCostProfileSpec(p []interface{}) (*costpb.CostProfileSpec, error) {
	obj := &costpb.CostProfileSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandCostProfileSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["provider_type"].(string); ok && len(v) > 0 {
		obj.ProviderType = v
	}

	if v, ok := in["installation_params"].([]interface{}); ok && len(v) > 0 {
		obj.InstallationParams = expandCostProfileIP(v)
	}

	return obj, nil
}

func expandCostProfileIP(p []interface{}) *costpb.InstallationParams {
	obj := &costpb.InstallationParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if _, ok := in["aws"]; ok {
		if v, ok := in["aws"].([]interface{}); ok && len(v) > 0 {
			obj.Aws = expandCostProfileAwsCostProfile(v)
		}
	}

	if _, ok := in["azure"]; ok {
		if v, ok := in["azure"].([]interface{}); ok && len(v) > 0 {
			obj.Azure = expandCostProfileAzureCostProfile(v)
		}
	}
	if _, ok := in["gcp"]; ok {
		if v, ok := in["gcp"].([]interface{}); ok && len(v) > 0 {
			obj.Gcp = expandCostProfileGcpCostProfile(v)
		}
	}

	if v, ok := in["other"].([]interface{}); ok && len(v) > 0 {
		obj.Other = expandCostProfileOtherCostProfile(v)
	}

	return obj

}
func expandCostProfileGcpCostProfile(p []interface{}) *costpb.GcpCostProfile {
	obj := &costpb.GcpCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["gcp_credentials"].([]interface{}); ok && len(v) > 0 {
		obj.GcpCredentials = expandCostProfileGcpCredentials(v)
	}
	return obj

}
func expandCostProfileAwsCostProfile(p []interface{}) *costpb.AwsCostProfile {
	obj := &costpb.AwsCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["aws_credentials"].([]interface{}); ok && len(v) > 0 {
		obj.AwsCredentials = expandCostProfileAwsCredentials(v)
	}

	if v, ok := in["cur_integration"].([]interface{}); ok && len(v) > 0 {
		obj.CurIntegration = expandCostProfileAwsCurIntegration(v)
	}

	if v, ok := in["spot_integration"].([]interface{}); ok && len(v) > 0 {
		obj.SpotIntegration = expandCostProfileAwsSpotIntegration(v)
	}

	return obj

}

func expandCostProfileAwsCredentials(p []interface{}) *costpb.AwsCredsCostProfile {
	obj := &costpb.AwsCredsCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["aws_service_key_name"].(string); ok && len(v) > 0 {
		obj.AwsServiceKeyName = v
	}
	if v, ok := in["aws_service_key_secret"].(string); ok && len(v) > 0 {
		obj.AwsServiceKeySecret = v
	}
	if v, ok := in["cloud_credentials_name"].(string); ok && len(v) > 0 {
		obj.CloudCredentialsName = v
	}
	if v, ok := in["role_arn"].(string); ok {
		obj.RoleArn = v
	}

	return obj

}

func expandCostProfileGcpCredentials(p []interface{}) *costpb.GcpCredsCostProfile{
	obj := &costpb.GcpCredsCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["cloud_credentials_name"].(string); ok && len(v) > 0 {
		obj.CloudCredentialsName = v
	}

	return obj
}

func expandCostProfileAwsCurIntegration(p []interface{}) *costpb.AwsCurIntegration {
	obj := &costpb.AwsCurIntegration{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["athena_bucket_name"].(string); ok && len(v) > 0 {
		obj.AthenaBucketName = v
	}
	if v, ok := in["athena_database"].(string); ok && len(v) > 0 {
		obj.AthenaDatabase = v
	}
	if v, ok := in["athena_region"].(string); ok && len(v) > 0 {
		obj.AthenaRegion = v
	}
	if v, ok := in["athena_table"].(string); ok && len(v) > 0 {
		obj.AthenaTable = v
	}
	if v, ok := in["aws_account_id"].(string); ok && len(v) > 0 {
		obj.AwsAccountId = v
	}
	if v, ok := in["master_payer_arn"].(string); ok && len(v) > 0 {
		obj.MasterPayerArn = v
	}

	return obj

}

func expandCostProfileAwsSpotIntegration(p []interface{}) *costpb.AwsSpotIntegration {
	obj := &costpb.AwsSpotIntegration{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["aws_account_id"].(string); ok && len(v) > 0 {
		obj.AwsAccountId = v
	}
	if v, ok := in["aws_spot_data_bucket"].(string); ok && len(v) > 0 {
		obj.AwsSpotDataBucket = v
	}
	if v, ok := in["aws_spot_data_prefix"].(string); ok && len(v) > 0 {
		obj.AwsSpotDataPrefix = v
	}
	if v, ok := in["aws_spot_data_region"].(string); ok && len(v) > 0 {
		obj.AwsSpotDataRegion = v
	}
	if v, ok := in["spot_label"].(string); ok && len(v) > 0 {
		obj.SpotLabel = v
	}
	if v, ok := in["spot_label_value"].(string); ok && len(v) > 0 {
		obj.SpotLabelValue = v
	}

	return obj

}

func expandCostProfileAzureCostProfile(p []interface{}) *costpb.AzureCostProfile {
	obj := &costpb.AzureCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["custom_pricing"].([]interface{}); ok && len(v) > 0 {
		obj.CustomPricing = expandCostProfileAzureCustomPricing(v)
	}

	if v, ok := in["gpu_estimates"].([]interface{}); ok && len(v) > 0 {
		obj.GpuEstimates = expandCostProfileAzureGpuEstimates(v)
	}

	return obj

}

func expandCostProfileAzureCustomPricing(p []interface{}) *costpb.AzureCustomPricing {
	obj := &costpb.AzureCustomPricing{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["azure_client_id"].(string); ok && len(v) > 0 {
		obj.AzureClientID = v
	}
	if v, ok := in["azure_client_secret"].(string); ok && len(v) > 0 {
		obj.AzureClientSecret = v
	}
	if v, ok := in["azure_subscription_id"].(string); ok && len(v) > 0 {
		obj.AzureSubscriptionID = v
	}
	if v, ok := in["azure_tenant_id"].(string); ok && len(v) > 0 {
		obj.AzureTenantID = v
	}
	if v, ok := in["cloud_credentials_name"].(string); ok && len(v) > 0 {
		obj.CloudCredentialsName = v
	}
	if v, ok := in["billing_account_id"].(string); ok && len(v) > 0 {
		obj.BillingAccountID = v
	}
	if v, ok := in["offer_id"].(string); ok && len(v) > 0 {
		obj.OfferID = v
	}
	if v, ok := in["spot_instance"].([]interface{}); ok && len(v) > 0 {
		obj.SpotInstance = expandCostProfileAzureSpotInstance(v)
	}

	return obj

}

func expandCostProfileAzureSpotInstance(p []interface{}) *costpb.SpotInstance {
	obj := &costpb.SpotInstance{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["spot_label"].(string); ok && len(v) > 0 {
		obj.SpotLabel = v
	}
	if v, ok := in["spot_label_value"].(string); ok && len(v) > 0 {
		obj.SpotLabelValue = v
	}

	return obj

}

func expandCostProfileAzureGpuEstimates(p []interface{}) *costpb.GpuCostProfile {
	obj := &costpb.GpuCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["gpu_label"].(string); ok && len(v) > 0 {
		obj.GpuLabel = v
	}
	if v, ok := in["gpu_label_value"].(string); ok && len(v) > 0 {
		obj.GpuLabelValue = v
	}

	return obj

}

func expandCostProfileOtherCostProfile(p []interface{}) *costpb.OtherCostProfile {
	obj := &costpb.OtherCostProfile{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cpu"].(string); ok && len(v) > 0 {
		obj.Cpu = v
	}

	if v, ok := in["gpu"].(string); ok && len(v) > 0 {
		obj.Gpu = v
	}

	if v, ok := in["memory"].(string); ok && len(v) > 0 {
		obj.Memory = v
	}

	return obj

}

// Flatteners

func flattenCostProfile(d *schema.ResourceData, in *costpb.CostProfile) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenCostProfileSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenCostProfileSpec(in *costpb.CostProfileSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenCostProfileSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	if len(in.ProviderType) > 0 {
		obj["provider_type"] = in.ProviderType
	}

	if in.InstallationParams != nil {
		v, ok := obj["installation_params"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["installation_params"] = flattenCostProfileSpecIP(in.InstallationParams, v)
	}

	return []interface{}{obj}, nil
}

func flattenCostProfileSpecIP(in *costpb.InstallationParams, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Aws != nil {
		v, ok := obj["aws"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["aws"] = flattenCostProfileAwsIP(in.Aws, v)
	}

	if in.Azure != nil {
		v, ok := obj["azure"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["azure"] = flattenCostProfileAzureIP(in.Azure, v)
	}

	if in.Other != nil {
		v, ok := obj["other"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["other"] = flattenCostProfileOtherIP(in.Other, v)
	}

	if in.Gcp != nil {
		v, ok := obj["gcp"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["gcp"] = flattenCostProfileGcpIP(in.Gcp, v)
	}

	return []interface{}{obj}
}

func flattenCostProfileAwsIP(in *costpb.AwsCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AwsCredentials != nil {
		v, ok := obj["aws_credentials"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["aws_credentials"] = flattenCostProfileAwsCredentials(in.AwsCredentials, v)
	}

	if in.CurIntegration != nil {
		v, ok := obj["cur_integration"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cur_integration"] = flattenCostProfileAwsCurIntegration(in.CurIntegration, v)
	}

	if in.SpotIntegration != nil {
		v, ok := obj["spot_integration"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["spot_integration"] = flattenCostProfileAwsSpotIntegration(in.SpotIntegration, v)
	}

	return []interface{}{obj}
}
func flattenCostProfileGcpIP(in *costpb.GcpCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.GcpCredentials != nil {
		v, ok := obj["gcp_credentials"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["gcp_credentials"] = flattenCostProfileGcpCredentials(in.GcpCredentials, v)
	}
	return []interface{}{obj}
}

func flattenCostProfileGcpCredentials(in *costpb.GcpCredsCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if len(in.CloudCredentialsName) > 0 {
		obj["cloud_credentials_name"] = in.CloudCredentialsName
	}
	return []interface{}{obj}

}
func flattenCostProfileAwsCredentials(in *costpb.AwsCredsCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.AwsServiceKeyName) > 0 {
		obj["aws_service_key_name"] = in.AwsServiceKeyName
	}

	if len(in.AwsServiceKeySecret) > 0 {
		obj["aws_service_key_secret"] = in.AwsServiceKeySecret
	}

	if len(in.CloudCredentialsName) > 0 {
		obj["cloud_credentials_name"] = in.CloudCredentialsName
	}

	if len(in.RoleArn) > 0 {
		obj["role_arn"] = in.RoleArn
	}

	return []interface{}{obj}
}

func flattenCostProfileAwsCurIntegration(in *costpb.AwsCurIntegration, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.AthenaBucketName) > 0 {
		obj["athena_bucket_name"] = in.AthenaBucketName
	}

	if len(in.AthenaDatabase) > 0 {
		obj["athena_database"] = in.AthenaDatabase
	}

	if len(in.AthenaRegion) > 0 {
		obj["athena_region"] = in.AthenaRegion
	}

	if len(in.AthenaTable) > 0 {
		obj["athena_table"] = in.AthenaTable
	}

	if len(in.AwsAccountId) > 0 {
		obj["aws_account_id"] = in.AwsAccountId
	}

	if len(in.MasterPayerArn) > 0 {
		obj["master_payer_arn"] = in.MasterPayerArn
	}

	return []interface{}{obj}
}

func flattenCostProfileAwsSpotIntegration(in *costpb.AwsSpotIntegration, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.AwsAccountId) > 0 {
		obj["aws_account_id"] = in.AwsAccountId
	}

	if len(in.AwsSpotDataBucket) > 0 {
		obj["aws_spot_data_bucket"] = in.AwsSpotDataBucket
	}

	if len(in.AwsSpotDataPrefix) > 0 {
		obj["aws_spot_data_prefix"] = in.AwsSpotDataPrefix
	}

	if len(in.AwsSpotDataRegion) > 0 {
		obj["aws_spot_data_region"] = in.AwsSpotDataRegion
	}

	if len(in.SpotLabel) > 0 {
		obj["spot_label"] = in.SpotLabel
	}

	if len(in.SpotLabelValue) > 0 {
		obj["spot_label_value"] = in.SpotLabelValue
	}

	return []interface{}{obj}
}

func flattenCostProfileAzureIP(in *costpb.AzureCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.CustomPricing != nil {
		v, ok := obj["custom_pricing"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["custom_pricing"] = flattenCostProfileAzureCustomPricing(in.CustomPricing, v)
	}

	if in.GpuEstimates != nil {
		v, ok := obj["gpu_estimates"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["gpu_estimates"] = flattenCostProfileAzureGpuEstimate(in.GpuEstimates, v)
	}

	return []interface{}{obj}
}

func flattenCostProfileAzureCustomPricing(in *costpb.AzureCustomPricing, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.AzureClientID) > 0 {
		obj["azure_client_id"] = in.AzureClientID
	}

	if len(in.AzureClientSecret) > 0 {
		obj["azure_client_secret"] = in.AzureClientSecret
	}

	if len(in.AzureSubscriptionID) > 0 {
		obj["azure_subscription_id"] = in.AzureSubscriptionID
	}

	if len(in.AzureTenantID) > 0 {
		obj["azure_tenant_id"] = in.AzureTenantID
	}

	if len(in.CloudCredentialsName) > 0 {
		obj["cloud_credentials_name"] = in.CloudCredentialsName
	}

	if len(in.BillingAccountID) > 0 {
		obj["billing_account_id"] = in.BillingAccountID
	}

	if len(in.OfferID) > 0 {
		obj["offer_id"] = in.OfferID
	}

	if in.SpotInstance != nil {
		v, ok := obj["spot_instance"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["spot_instance"] = flattenCostProfileAzureSpotInstance(in.SpotInstance, v)
	}

	return []interface{}{obj}
}

func flattenCostProfileAzureSpotInstance(in *costpb.SpotInstance, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.SpotLabel) > 0 {
		obj["spot_label"] = in.SpotLabel
	}

	if len(in.SpotLabelValue) > 0 {
		obj["spot_label_value"] = in.SpotLabelValue
	}

	return []interface{}{obj}
}

func flattenCostProfileAzureGpuEstimate(in *costpb.GpuCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.GpuLabel) > 0 {
		obj["gpu_label"] = in.GpuLabel
	}

	if len(in.GpuLabelValue) > 0 {
		obj["gpu_label_value"] = in.GpuLabelValue
	}

	return []interface{}{obj}
}

func flattenCostProfileOtherIP(in *costpb.OtherCostProfile, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Cpu) > 0 {
		obj["cpu"] = in.Cpu
	}

	if len(in.Gpu) > 0 {
		obj["gpu"] = in.Gpu
	}

	if len(in.Memory) > 0 {
		obj["memory"] = in.Memory
	}

	return []interface{}{obj}
}
