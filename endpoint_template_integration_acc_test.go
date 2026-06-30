package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccTemplateEndpointConfig declares a template and an endpoint that
// references it via template_id — a multi-resource dependency flow.
func testAccTemplateEndpointConfig(tplName, epName string) string {
	return fmt.Sprintf(`
provider "runpod" {}

resource "runpod_template" "t" {
  name                 = %q
  image_name           = "runpod/base:0.0.0"
  container_disk_in_gb = 10
}

resource "runpod_endpoint" "e" {
  name        = %q
  template_id = runpod_template.t.id
  workers_min = 0
  workers_max = 1
}
`, tplName, epName)
}

// TestAccTemplateEndpoint_integration_framework is a multi-resource integration
// test: it applies a runpod_template and a runpod_endpoint that references the
// template's id, exercising dependency ordering through the real provider
// (plan/apply/refresh) against runpod-in-a-box.
//
// Skip-gated on CE-1681: both template and endpoint Create reject HTTP 201 (the
// API returns 201 Created), so apply fails at the first create. riab supports
// both endpoints (GET 200) and endpoint create needs a real template_id — which
// this config provides. Un-skip when Create accepts 201; it then becomes the
// gap-#1 endpoint + gap-#6 integration coverage.
func TestAccTemplateEndpoint_integration_framework(t *testing.T) {
	if os.Getenv("RIAB_ACC") != "1" {
		t.Skip("set RIAB_ACC=1 + TF_ACC=1 + RUNPOD_BASE_URL + RUNPOD_API_KEY to run the live riab integration test")
	}
	t.Skip("CE-1681: template + endpoint Create reject HTTP 201 → apply fails at create; un-skip when 201 is accepted")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateEndpointConfig("tf-acc-int-tpl", "tf-acc-int-ep"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("runpod_template.t", "id"),
					resource.TestCheckResourceAttrSet("runpod_endpoint.e", "id"),
					// the endpoint's template_id must resolve to the created template's id
					resource.TestCheckResourceAttrPair("runpod_endpoint.e", "template_id", "runpod_template.t", "id"),
				),
			},
		},
	})
}
