package provider

import (
	"context"
	"testing"

	"github.com/SSHcom/privx-sdk-go/v2/api/hoststore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestHostDataSource_populateModel verifies, without a live PrivX, that a host
// returned by the API is mapped into the data source model and can be written
// to state using the real schema (id, nested services, principals, etc.).
func TestHostDataSource_populateModel(t *testing.T) {
	ctx := context.Background()
	d := &HostDataSource{}

	var sresp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &sresp)

	host := &hoststore.Host{
		ID:            "3f2b8c1e-5a7d-4e9a-9b1c-2d6f8a0e4c11",
		CommonName:    "ip-172-31-14-176.eu-west-2.compute.internal",
		Addresses:     []string{"172.31.14.176"},
		AccessGroupID: "ag-1",
		Tags:          []string{"b", "a"},
		Services: []hoststore.HostService{
			{Service: "SSH", Address: "172.31.14.176", Port: 22, HealthCheckStatus: "OK"},
		},
		Principals: []hoststore.HostPrincipals{
			{Principal: "root", Rotate: true},
		},
	}

	data := HostDataSourceModel{
		CommonName: types.StringValue(host.CommonName),
	}

	var resp datasource.ReadResponse
	resp.State = tfsdk.State{Schema: sresp.Schema}

	d.populateModel(ctx, &data, host, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		t.Fatalf("populateModel returned errors: %v", resp.Diagnostics)
	}

	if got := data.ID.ValueString(); got != host.ID {
		t.Fatalf("id = %q, want %q", got, host.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("State.Set failed against the data source schema: %v", resp.Diagnostics)
	}

	if len(data.Services.Elements()) != 1 {
		t.Fatalf("services = %d, want 1", len(data.Services.Elements()))
	}
	if len(data.Principals.Elements()) != 1 {
		t.Fatalf("principals = %d, want 1", len(data.Principals.Elements()))
	}
}
