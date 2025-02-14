package pipeline

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func pipelineSlackProjectIntegrationResource() *schema.Resource {

	var slackIntegrationSchema = mergeSchema(
		projectIntegrationSchema,
		map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "url for Slack access",
			},
			"master_integration_id": {
				Type:        schema.TypeInt,
				Default:     78,
				Optional:    true,
				Description: "The Id of the master integration.",
			},
			"master_integration_name": {
				Type:        schema.TypeString,
				Default:     "slackKey",
				Optional:    true,
				Description: "The name of the master integration.",
			},
		},
	)

	var unpackSlackFormValues = func(data *schema.ResourceData) []FormJSONValues {
		d := &ResourceData{data}
		var formJSONValues = []FormJSONValues{
			{
				Label: "url",
				Value: d.getString("url"),
			},
		}
		return formJSONValues
	}

	var readSlackProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] readSlackProjectIntegration")
		_, err := readProjectIntegration(data, m)
		if err != nil {
			return diag.FromErr(err)
		}

		return nil
	}

	var createSlackProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] createSlackProjectIntegration")

		slackFormValues := unpackSlackFormValues(data)
		err := createProjectIntegration(data, m, slackFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readSlackProjectIntegration(ctx, data, m)
	}

	var updateSlackProjectIntegration = func(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
		log.Printf("[DEBUG] updateSlackProjectIntegration")

		slackFormValues := unpackSlackFormValues(data)
		err := updateProjectIntegration(data, m, slackFormValues)
		if err != nil {
			return diag.FromErr(err)
		}

		return readSlackProjectIntegration(ctx, data, m)
	}

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: createSlackProjectIntegration,
		ReadContext:   readSlackProjectIntegration,
		UpdateContext: updateSlackProjectIntegration,
		DeleteContext: deleteProjectIntegration,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      slackIntegrationSchema,
		Description: "Provides an Jfrog Pipelines Slack Project Integration resource.",
	}
}
