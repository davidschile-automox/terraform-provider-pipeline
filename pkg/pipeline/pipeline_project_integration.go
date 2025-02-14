package pipeline

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ProjectIntegration GET {{ host }}/pipelines/api/v1/projectintegrations/{{projectIntegrationId}}

type ProjectIntegration struct {
	Name                  string           `json:"name"`
	ProjectId             int              `json:"projectId,omitempty"`
	Project               ProjectJSON      `json:"project,omitempty"`
	MasterIntegrationId   int              `json:"masterIntegrationId"`
	MasterIntegrationName string           `json:"masterIntegrationName"`
	FormJSONValues        []FormJSONValues `json:"formJSONValues"`
	Environments          []string         `json:"environments,omitempty"`
	IsInternal            bool             `json:"isInternal,omitempty"`
	ID                    int              `json:"id,omitempty"`
}

type FormJSONValues struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ProjectJSON struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

const projectIntegrationsUrl = "pipelines/api/v1/projectintegrations"

var projectIntegrationSchema = map[string]*schema.Schema{
	"name": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotEmpty,
		Description:  "The name of the project integration. Should be prefixed with the project key",
	},

	"project_id": {
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Description:  "Id of the project.",
	},
	"project": {
		Type:     schema.TypeMap,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "An object containing a project name as an alternative to projectId.",
	},
	"environments": {
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "In a project, an array of environment names in which this pipeline source will be.",
	},
	"is_internal": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Set this as false to create a Pipelines integration.",
	},
}

func unpackProject(d *ResourceData, key string) ProjectJSON {
	var project ProjectJSON
	input := d.Get(key).(map[string]interface{})
	project.Key = input["key"].(string)
	project.Name = input["name"].(string)

	return project
}

func packProject(d *schema.ResourceData, schemaKey string, project ProjectJSON) []error {
	var errors []error
	log.Println("[DEBUG] packProject", project)
	if (ProjectJSON{}) == project {
		return errors
	}
	setValue := mkLens(d)
	errors = append(errors, setValue(schemaKey, project)...)
	return errors
}

func unpackProjectIntegration(data *schema.ResourceData, formJsonValues []FormJSONValues) (ProjectIntegration, error) {
	d := &ResourceData{data}

	projectIntegration := ProjectIntegration{
		Name:                  d.getString("name"),
		ProjectId:             d.getInt("project_id"),
		MasterIntegrationId:   d.getInt("master_integration_id"),
		MasterIntegrationName: d.getString("master_integration_name"),
		Environments:          d.getSet("environments"),
		IsInternal:            d.getBool("is_internal"),
		Project:               unpackProject(d, "project"),
		FormJSONValues:        formJsonValues,
	}
	return projectIntegration, nil
}

func packProjectIntegration(d *schema.ResourceData, projectIntegration ProjectIntegration) []error {
	var errors []error
	setValue := mkLens(d)

	errors = setValue("name", projectIntegration.Name)
	errors = append(errors, setValue("master_integration_id", projectIntegration.MasterIntegrationId)...)
	errors = append(errors, setValue("master_integration_name", projectIntegration.MasterIntegrationName)...)
	errors = append(errors, setValue("environments", projectIntegration.Environments)...)
	errors = append(errors, setValue("is_internal", projectIntegration.IsInternal)...)
	errors = append(errors, packProject(d, "project", projectIntegration.Project)...)

	return errors
}

func readProjectIntegration(data *schema.ResourceData, m interface{}) ([]FormJSONValues, error) {
	log.Printf("[DEBUG] readProjectIntegration")
	projectIntegration := ProjectIntegration{}
	resp, err := m.(*resty.Client).R().
		SetResult(&projectIntegration).
		Get(projectIntegrationsUrl + "/" + data.Id())
	log.Println("[DEBUG] projectIntegration body: ", string(json.RawMessage(resp.Body())))
	if err != nil {
		return nil, err
	}
	log.Println("[DEBUG] projectIntegration Obj: ", projectIntegration)
	errors := packProjectIntegration(data, projectIntegration)
	if len(errors) > 0 {
		return nil, errors[0]
	}
	return projectIntegration.FormJSONValues, nil
}

func createProjectIntegration(data *schema.ResourceData, m interface{}, formValues []FormJSONValues) error {
	log.Printf("[DEBUG] createProjectIntegration")
	log.Printf("[TRACE] %+v\n", data)

	projectIntegration, err := unpackProjectIntegration(data, formValues)
	if err != nil {
		return err
	}
	response := ProjectIntegration{}
	_, err = m.(*resty.Client).R().
		SetBody(projectIntegration).
		SetResult(&response).
		Post(projectIntegrationsUrl)
	if err != nil {
		return err
	}
	data.SetId(strconv.Itoa(response.ID))

	return nil
}

func updateProjectIntegration(data *schema.ResourceData, m interface{}, formValues []FormJSONValues) error {
	log.Printf("[DEBUG] updateProjectIntegration")
	log.Printf("[TRACE] %+v\n", data)

	projectIntegration, err := unpackProjectIntegration(data, formValues)
	if err != nil {
		return err
	}

	_, err = m.(*resty.Client).R().SetBody(projectIntegration).Put(projectIntegrationsUrl + "/" + data.Id())
	if err != nil {
		return err
	}

	return nil
}

func deleteProjectIntegration(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] deleteProjectIntegration")
	log.Printf("[TRACE] %+v\n", data)

	resp, err := m.(*resty.Client).R().
		Delete(projectIntegrationsUrl + "/" + data.Id())

	if err != nil && resp.StatusCode() == http.StatusNotFound {
		data.SetId("")
		return diag.FromErr(err)
	}

	return nil
}
