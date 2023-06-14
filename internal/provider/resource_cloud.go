package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jujucloud "github.com/juju/juju/cloud"
	"github.com/juju/juju/rpc/params"
	"github.com/juju/terraform-provider-juju/internal/juju"
)

func resourceCloud() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A resource that represent a Juju Cloud.",

		CreateContext: resourceCloudCreate,
		ReadContext:   resourceCloudRead,
		UpdateContext: resourceCloudUpdate,
		DeleteContext: resourceCloudDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name to be assigned to the cloud",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description: "The type of cloud (e.g. ec2, openstack, lxd)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"host_cloud_region": {
				Description: "Represents the k8s host cloud. The format is <cloudType>/<region>",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"auth_types": {
				Description: "The authentication modes supported by the cloud",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"endpoint": {
				Description: "The default endpoint for the cloud regions",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"identity_endpoint": {
				Description: "The default identity endpoint for the cloud regions",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"storage_endpoint": {
				Description: "The default storage endpoint for the cloud regions",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"regions": {
				Description: "Regions available within the cloud",
				Type:        schema.TypeSet,
				Optional:    true,
				// Set as "computed" to pre-populate and preserve any implicit values
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the region",
							Type:        schema.TypeString,
							Required:    true,
						},
						"endpoint": {
							Description: "The endpoint URL for this region",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"identity_endpoint": {
							Description: "The identity endpoint URL for this region",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"storage_endpoint": {
							Description: "The storage endpoint URL for this region",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"config": {
				Description: "Optional cloud-specific configuration",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"region_config": {
				Description: "Optional region-specific configuration",
				Type:        schema.TypeSet,
				Optional:    true,
				// Set as "computed" to pre-populate and preserve any implicit values
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Name of the region config applies to",
							Type:        schema.TypeString,
							Required:    true,
						},
						"config": {
							Description: "Config applied to region",
							Type:        schema.TypeMap,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"ca_certificates": {
				Description: "List of CA certs to be used to validate certificates of cloud infrastructure components.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"skip_tls_verify": {
				Description: "Skip certificate validation. Not recommended for production clouds.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"is_controller_cloud": {
				Description: "True when this is the cloud used by the controller",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func resourceCloudCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*juju.Client)

	var diags diag.Diagnostics

	cloudParams := prepareCloudInput(d)

	var cloud juju.CreateCloudInput
	cloud.Name = d.Get("name").(string)
	cloud.Params = cloudParams

	err := client.Clouds.CreateCloud(cloud)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("cloud:%s", cloud.Name))

	resourceCloudRead(ctx, d, meta)

	return diags
}

func resourceCloudRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*juju.Client)

	var diags diag.Diagnostics

	id := strings.Split(d.Id(), ":")
	name := id[1]
	response, err := client.Clouds.ReadCloud(juju.ReadCloudInput{
		Name: name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", response.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", response.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("host_cloud_region", response.HostCloudRegion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_types", response.AuthTypes); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("endpoint", response.Endpoint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("identity_endpoint", response.IdentityEndpoint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("storage_endpoint", response.StorageEndpoint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("config", response.Config); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ca_certificates", response.CACertificates); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("skip_tls_verify", response.SkipTLSVerify); err != nil {
		return diag.FromErr(err)
	}

	regions := parseRegions(response.Regions)
	if err := d.Set("regions", regions); err != nil {
		return diag.FromErr(err)
	}

	regionConfig := parseRegionConfig(response.RegionConfig)
	if err := d.Set("region_config", regionConfig); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*juju.Client)

	var diags diag.Diagnostics
	anyChange := false

	var err error

	if d.HasChange("host_cloud_region") {
		anyChange = true
	}
	if d.HasChange("auth_types") {
		anyChange = true
	}
	if d.HasChange("endpoint") {
		anyChange = true
	}
	if d.HasChange("identity_endpoint") {
		anyChange = true
	}
	if d.HasChange("storage_endpoint") {
		anyChange = true
	}
	if d.HasChange("config") {
		anyChange = true
	}
	if d.HasChange("ca_certificates") {
		anyChange = true
	}
	if d.HasChange("skip_tls_verify") {
		anyChange = true
	}
	if d.HasChange("regions") {
		anyChange = true
	}
	if d.HasChange("region_config") {
		anyChange = true
	}

	if !anyChange {
		return diags
	}

	cloudParams := prepareCloudInput(d)

	var cloud juju.UpdateCloudInput
	cloud.Name = d.Get("name").(string)
	cloud.Params = cloudParams

	err = client.Clouds.UpdateCloud(cloud)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*juju.Client)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	err := client.Clouds.RemoveCloud(juju.RemoveCloudInput{
		Name: name,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func parseRegions(regions []jujucloud.Region) []map[string]interface{} {
	cloudRegions := make([]map[string]interface{}, 0)

	for _, region := range regions {
		r := make(map[string]interface{})
		r["name"] = region.Name
		if region.Endpoint != "" {
			r["endpoint"] = region.Endpoint
		}
		if region.IdentityEndpoint != "" {
			r["identity_endpoint"] = region.IdentityEndpoint
		}
		if region.StorageEndpoint != "" {
			r["storage_endpoint"] = region.StorageEndpoint
		}
		cloudRegions = append(cloudRegions, r)
	}
	return cloudRegions
}

func parseRegionConfig(regionConfig jujucloud.RegionConfig) []map[string]interface{} {
	cloudRegionConfig := make([]map[string]interface{}, 0)

	for region, config := range regionConfig {
		r := make(map[string]interface{})
		r["name"] = region
		r["config"] = config
		println(r["name"])
		cloudRegionConfig = append(cloudRegionConfig, r)
	}
	return cloudRegionConfig
}

func prepareCloudInput(d *schema.ResourceData) params.Cloud {
	var cloudParams params.Cloud

	cloudParams.Type = d.Get("type").(string)

	hostCloudRegion := d.Get("host_cloud_region").(string)
	if hostCloudRegion != "" {
		cloudParams.HostCloudRegion = hostCloudRegion
	}

	authTypes := d.Get("auth_types").([]interface{})
	cloudAuthTypes := make([]string, len(authTypes))
	for i, authType := range authTypes {
		cloudAuthTypes[i] = authType.(string)
	}
	cloudParams.AuthTypes = cloudAuthTypes

	endpoint := d.Get("endpoint").(string)
	if endpoint != "" {
		cloudParams.Endpoint = endpoint
	}

	identityEndpoint := d.Get("identity_endpoint").(string)
	if identityEndpoint != "" {
		cloudParams.IdentityEndpoint = identityEndpoint
	}

	storageEndpoint := d.Get("storage_endpoint").(string)
	if storageEndpoint != "" {
		cloudParams.StorageEndpoint = storageEndpoint
	}

	regions := d.Get("regions").(*schema.Set).List()
	if len(regions) > 0 {
		cloudRegions := make([]params.CloudRegion, 0)
		for _, r := range regions {
			region := r.(map[string]interface{})
			cloudRegion := params.CloudRegion{
				Name: region["name"].(string),
			}
			if val, ok := region["endpoint"].(string); ok {
				cloudRegion.Endpoint = val
			}
			if val, ok := region["identity_endpoint"].(string); ok {
				cloudRegion.IdentityEndpoint = val
			}
			if val, ok := region["identity_endpoint"].(string); ok {
				cloudRegion.IdentityEndpoint = val
			}
		}
		cloudParams.Regions = cloudRegions
	}

	regionConfig := d.Get("region_config").(*schema.Set).List()
	if len(regionConfig) > 0 {
		var cloudRegionConfig map[string]map[string]interface{}
		for _, r := range regionConfig {
			region := r.(map[string]interface{})
			cloudRegionConfig[region["name"].(string)] = region["config"].(map[string]interface{})
		}
		cloudParams.RegionConfig = cloudRegionConfig
	}

	configField := d.Get("config").(map[string]interface{})
	if len(configField) > 0 {
		cloudParams.Config = configField
	}

	caCertificates := d.Get("ca_certificates").([]interface{})
	if len(caCertificates) > 0 {
		cloudCACertificates := make([]string, len(caCertificates))
		for i, cert := range caCertificates {
			cloudCACertificates[i] = cert.(string)
		}
		cloudParams.CACertificates = cloudCACertificates
	}

	cloudParams.SkipTLSVerify = d.Get("skip_tls_verify").(bool)
	return cloudParams
}
