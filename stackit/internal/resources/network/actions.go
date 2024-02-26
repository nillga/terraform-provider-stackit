package network

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/SchwarzIT/community-stackit-go-client/pkg/validate"
	"github.com/SchwarzIT/terraform-provider-stackit/stackit/internal/common"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	iaas "github.com/SchwarzIT/community-stackit-go-client/pkg/services/iaas-api/v1alpha"
	clientValidate "github.com/SchwarzIT/community-stackit-go-client/pkg/validate"
)

// Create - lifecycle function
func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Network
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// handle creation
	plan = r.createNetwork(ctx, resp, plan)
	if resp.Diagnostics.HasError() {
		return
	}

	// update state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r Resource) createNetwork(ctx context.Context, resp *resource.CreateResponse, plan Network) Network {
	var ns iaas.V1Nameserver

	for _, i := range plan.NameServers.Elements() {
		if i.IsNull() || i.IsUnknown() {
			continue
		}

		nsVal, err := common.ToString(context.TODO(), i)
		if err != nil {
			continue
		}

		ns = append(ns, nsVal)
	}

	pl := int(plan.PrefixLengthV4.ValueInt64())
	name := plan.Name.ValueString()

	body := iaas.V1CreateNetworkJSONRequestBody{
		Name:           name,
		Nameservers:    &ns,
		PrefixLengthV4: &pl,
	}

	projectID, _ := uuid.Parse(plan.ProjectID.String())

	res, err := r.client.IAAS.V1CreateNetwork(ctx, projectID, body)

	timeout, d := plan.Timeouts.Create(ctx, 5*time.Minute)
	if resp.Diagnostics.Append(d...); resp.Diagnostics.HasError() {
		return plan
	}

	process := res.WaitHandler(ctx, r.client.IAAS, projectID, name).SetTimeout(timeout)
	wr, err := process.WaitWithContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed validating network %s creation", body.Name), err.Error())
		return plan
	}

	network, ok := wr.(iaas.V1Network)
	if !ok {
		resp.Diagnostics.AddError("failed wait result conversion", "result is not of *iaas.V1Network")
		return plan
	}

	prefixes := make([]attr.Value, 0)

	if len(network.Prefixes) > 0 {
		for _, pr := range network.Prefixes {
			prefixes = append(prefixes, types.StringValue(pr))
		}
	}

	plan.ID = types.StringValue(network.NetworkID.String())
	plan.NetworkID = types.StringValue(network.NetworkID.String())
	plan.PublicIp = types.StringPointerValue(network.PublicIp)
	plan.Prefixes = types.ListValueMust(types.StringType, prefixes)
	plan.ProjectID = types.StringValue(projectID.String())

	return plan
}

// Read - lifecycle function
func (r Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	c := r.client
	var state Network

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, _ := uuid.Parse(state.ProjectID.ValueString())
	networkID, _ := uuid.Parse(state.NetworkID.ValueString())

	res, err := c.IAAS.V1GetNetwork(ctx, projectID, networkID)
	if agg := common.Validate(&resp.Diagnostics, res, err, "JSON200"); agg != nil {
		if validate.StatusEquals(res, http.StatusNotFound) {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("failed reading project", agg.Error())
		return
	}

	n := res.JSON200

	prefixes := make([]attr.Value, 0)
	if len(n.Prefixes) > 0 {
		for _, pr := range n.Prefixes {
			prefixes = append(prefixes, types.StringValue(pr))
		}
	}

	nameservers := make([]attr.Value, 0)
	if n.Nameservers != nil && len(*n.Nameservers) > 0 {
		for _, ns := range *n.Nameservers {
			nameservers = append(nameservers, types.StringValue(ns))
		}
	}

	state.ID = types.StringValue(n.NetworkID.String())
	state.ProjectID = types.StringValue(projectID.String())
	state.Name = types.StringValue(n.Name)
	state.NetworkID = types.StringValue(n.NetworkID.String())
	state.PublicIp = types.StringPointerValue(n.PublicIp)
	state.Prefixes = types.ListValueMust(types.StringType, prefixes)
	state.NameServers = types.ListValueMust(types.StringType, nameservers)

	// get the Prefix Length in a hacky way, otherwise fall back to default
	if len(n.Prefixes) > 0 {
		cidrSplit := strings.Split(n.Prefixes[0], "/")
		if len(cidrSplit) != 2 {
			resp.Diagnostics.AddError("Processing CIDR Prefix Length",
				"Processing CIDR Prefix Length")
			return
		}

		prefixLength, err := strconv.ParseInt(cidrSplit[1], 10, 64)
		if err != nil {
			resp.Diagnostics.AddError("Processing CIDR Prefix Length", err.Error())
			return
		}

		state.PrefixLengthV4 = types.Int64Value(prefixLength)
	} else {
		state.PrefixLengthV4 = types.Int64Value(25)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update - lifecycle function
func (r Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Network
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Network
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.NetworkID.IsUnknown() {
		plan.NetworkID = state.NetworkID
	}

	r.updateNetwork(ctx, plan, state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// update state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r Resource) updateNetwork(ctx context.Context, plan, state Network, resp *resource.UpdateResponse) {
	if plan.Name.Equal(state.Name) && reflect.DeepEqual(plan.NameServers, state.NameServers) {
		return
	}

	ns := make([]iaas.V1IP, 0)
	for _, s := range plan.NameServers.Elements() {
		if s.IsNull() || s.IsUnknown() {
			continue
		}

		ns = append(ns, s.String())
	}

	name := plan.Name.ValueString()

	body := iaas.V1UpdateNetworkJSONBody{
		Name:        &name,
		Nameservers: &ns,
	}

	projectID, _ := uuid.Parse(state.ProjectID.String())
	networkID, _ := uuid.Parse(state.NetworkID.String())

	res, err := r.client.IAAS.V1UpdateNetwork(ctx, projectID, networkID, iaas.V1UpdateNetworkJSONRequestBody(body))
	if agg := common.Validate(&resp.Diagnostics, res, err, "JSON200"); agg != nil {
		resp.Diagnostics.AddError("failed updating project", agg.Error())
		return
	}
}

// Delete - lifecycle function
func (r Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Network
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, _ := uuid.Parse(state.ProjectID.String())
	networkID, _ := uuid.Parse(state.NetworkID.String())
	c := r.client
	res, err := c.IAAS.V1DeleteNetwork(ctx, projectID, networkID)
	if agg := common.Validate(&resp.Diagnostics, res, err); agg != nil {
		resp.Diagnostics.AddError("failed deleting network", agg.Error())
		return
	}

	process := res.WaitHandler(ctx, c.IAAS, projectID, networkID)
	if _, err = process.WaitWithContext(ctx); err != nil {
		resp.Diagnostics.AddError("failed to verify network deletion", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

// ImportState handles terraform import
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: `project_id,id` where `id` is the network_id and `project_id` is the project id.\nInstead got: %q", req.ID),
		)
		return
	}

	projectID := idParts[0]
	networkID := idParts[1]

	if err := clientValidate.ProjectID(projectID); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Couldn't validate project_id.\n%s", err.Error()),
		)
		return
	}
	if err := clientValidate.NetworkID(networkID); err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Couldn't validate network_id.\n%s", err.Error()),
		)
		return
	}

	// set main attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_id"), networkID)...)
}
