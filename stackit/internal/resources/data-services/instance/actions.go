package instance

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SchwarzIT/community-stackit-go-client/pkg/services/data-services/v1.0/generated/instances"
	clientValidate "github.com/SchwarzIT/community-stackit-go-client/pkg/validate"
	"github.com/SchwarzIT/terraform-provider-stackit/stackit/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Create - lifecycle function
func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// load plan
	var plan Instance
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validate
	if err := r.validate(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("failed instance validation", err.Error())
		return
	}

	acl := []string{}
	for _, v := range plan.ACL.Elements() {
		nv, err := common.ToString(ctx, v)
		if err != nil {
			continue
		}
		acl = append(acl, nv)
	}

	// handle creation
	params := map[string]interface{}{
		"sgw_acl": strings.Join(acl, ","),
	}
	body := instances.InstanceProvisionRequest{
		InstanceName: plan.Name.ValueString(),
		PlanID:       plan.PlanID.ValueString(),
		Parameters:   &params,
	}
	res, err := r.client.Instances.ProvisionWithResponse(ctx, plan.ProjectID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("failed preparing instance provisioning request", err.Error())
		return
	}
	if res.HasError != nil {
		resp.Diagnostics.AddError("failed making instance provisioning request", res.HasError.Error())
		return
	}
	if res.JSON202 == nil {
		resp.Diagnostics.AddError("received an empty response", "JSON202 == nil")
		return
	}

	// set state
	plan.ID = types.StringValue(res.JSON202.InstanceID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), res.JSON202.InstanceID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	process := res.WaitHandler(ctx, r.client.Instances, plan.ProjectID.ValueString(), plan.ID.ValueString())
	instance, err := process.WaitWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed instance `create` validation", err.Error())
		return
	}

	i, ok := instance.(*instances.GetResponse)
	if !ok {
		resp.Diagnostics.AddError("failed to parse client response", "response is not of *instances.GetResponse")
		return
	}

	if i.JSON200 == nil {
		resp.Diagnostics.AddError("failed to parse client response", "JSON200 == nil")
		return
	}

	if err := r.applyClientResponse(ctx, &plan, i.JSON200); err != nil {
		resp.Diagnostics.AddError("failed to process client response", err.Error())
		return
	}

	// update state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read - lifecycle function
func (r Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Instance
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read instance
	res, err := r.client.Instances.GetWithResponse(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed preparing get instance request", err.Error())
		return
	}
	if res.HasError != nil {
		if res.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed making get instance request", res.HasError.Error())
		return
	}
	if res.JSON200 == nil {
		resp.Diagnostics.AddError("received an empty response", "JSON200 == nil")
		return
	}

	if err := r.applyClientResponse(ctx, &state, res.JSON200); err != nil {
		resp.Diagnostics.AddError("failed to process client response", err.Error())
		return
	}

	// update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update - lifecycle function
func (r Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Instance
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	if plan.ACL.IsUnknown() || plan.ACL.IsNull() {
		plan.ACL = state.ACL
	}

	// validate
	if err := r.validate(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("failed validation", err.Error())
		return
	}

	acl := []string{}
	for _, v := range plan.ACL.Elements() {
		nv, err := common.ToString(ctx, v)
		if err != nil {
			continue
		}
		acl = append(acl, nv)
	}

	// handle update
	params := map[string]interface{}{
		"sgw_acl": strings.Join(acl, ","),
	}
	body := instances.UpdateJSONRequestBody{
		PlanID:     plan.PlanID.ValueString(),
		Parameters: &params,
	}
	res, err := r.client.Instances.UpdateWithResponse(ctx, state.ProjectID.ValueString(), state.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("failed preparing update instance request", err.Error())
		return
	}
	if res.HasError != nil {
		resp.Diagnostics.AddError("failed making update instance request", res.HasError.Error())
		return
	}

	process := res.WaitHandler(ctx, r.client.Instances, state.ProjectID.ValueString(), state.ID.ValueString())
	instancesGetResp, err := process.WaitWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed instance update validation", err.Error())
		return
	}

	newRes, ok := instancesGetResp.(*instances.GetResponse)
	if !ok {
		resp.Diagnostics.AddError("can't parse response", "returned value is not of *instances.GetResponse")
		return
	}
	if newRes.JSON200 == nil {
		resp.Diagnostics.AddError("received an empty response", "JSON200 == nil")
		return
	}

	planID := plan.PlanID
	if err := r.applyClientResponse(ctx, &plan, newRes.JSON200); err != nil {
		resp.Diagnostics.AddError("failed to process client response", err.Error())
		return
	}

	if !plan.PlanID.Equal(planID) {
		resp.Diagnostics.AddError("server returned wrong plan ID after update", fmt.Sprintf("expected plan ID %s but received %s", planID.ValueString(), plan.PlanID.ValueString()))
		return
	}

	// update state
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete - lifecycle function
func (r Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Instance
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// handle deletion
	res, err := r.client.Instances.DeprovisionWithResponse(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed making deprovision request", err.Error())
		return
	}
	if res.HasError != nil {
		if res.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed making deprovision instance request", res.HasError.Error())
		return
	}

	process := res.WaitHandler(ctx, r.client.Instances, state.ProjectID.ValueString(), state.ID.ValueString())
	if _, err := process.WaitWithContext(ctx); err != nil {
		resp.Diagnostics.AddError("failed to verify instance deprovision", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

// ImportState handles terraform import
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: `project_id,instance_id`.\nInstead got: %q", req.ID),
		)
		return
	}

	// validate project id
	if err := clientValidate.ProjectID(idParts[0]); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Couldn't validate project_id.\n%s", err.Error()),
		)
		return
	}

	plan, version, err := r.getPlanAndVersion(ctx, idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error during import",
			err.Error(),
		)
		return
	}
	// set main attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("plan"), plan)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version)...)

	if resp.Diagnostics.HasError() {
		return
	}
}