package rafay

import (
	"context"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Test helper functions to expose internal functions for testing

func TestResourceFleetPlan() *schema.Resource {
	return resourceFleetPlan()
}

func TestResourceFleetPlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanCreate(ctx, d, m)
}

func TestResourceFleetPlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanUpdate(ctx, d, m)
}

func TestResourceFleetPlanUpsert(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	return resourceFleetPlanUpsert(ctx, d)
}

func TestResourceFleetPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanRead(ctx, d, m)
}

func TestResourceFleetPlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanDelete(ctx, d, m)
}

func TestExpandFleetPlan(d *schema.ResourceData) (*infrapb.FleetPlan, error) {
	return expandFleetPlan(d)
}

func TestExpandFleetPlanSpec(p []interface{}) (*infrapb.FleetPlanSpec, error) {
	return expandFleetPlanSpec(p)
}

func TestFlattenFleetPlanSpec(spec *infrapb.FleetPlanSpec) []interface{} {
	return flattenFleetPlanSpec(spec)
}

func TestExpandFleetSpec(p []interface{}) *infrapb.FleetSpec {
	return expandFleetSpec(p)
}

func TestExpandProjects(v []interface{}) []*infrapb.ProjectFilter {
	return expandProjects(v)
}

func TestExpandEnvironmentTemplates(v []interface{}) []*infrapb.TemplateFilter {
	return expandEnvironmentTemplates(v)
}

func TestExpandFleetPlanSchedules(p []interface{}) *infrapb.FleetSchedule {
	return expandFleetPlanSchedules(p)
}

func TestExpandScheduleCadence(p []interface{}) *infrapb.ScheduleOptions {
	return expandScheduleCadence(p)
}

func TestFlattenScheduleCadence(cadence *infrapb.ScheduleOptions) []interface{} {
	return flattenScheduleCadence(cadence)
}

func TestFlattenSchedule(schedule *infrapb.FleetSchedule) []interface{} {
	return flattenSchedule(schedule)
}

func TestFlattenFleetPlanSchedules(schedules []*infrapb.FleetSchedule) []interface{} {
	return flattenFleetPlanSchedules(schedules)
}

func TestFlattenFleet(fs *infrapb.FleetSpec) []interface{} {
	return flattenFleet(fs)
}

func TestFlattenTemplates(templates []*infrapb.TemplateFilter) []interface{} {
	return flattenTemplates(templates)
}

func TestFlattenProjects(projects []*infrapb.ProjectFilter) []interface{} {
	return flattenProjects(projects)
}

// FleetPlan Job test helper functions

func TestResourceFleetPlanTrigger() *schema.Resource {
	return resourceFleetPlanTrigger()
}

func TestCreateFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return createFleetPlanJob(ctx, d, m)
}

func TestUpdateFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return updateFleetPlanJob(ctx, d, m)
}

func TestUpsertFleetPlanJob(ctx context.Context, d *schema.ResourceData) diag.Diagnostics {
	return upsertFleetPlanJob(ctx, d)
}

func TestReadFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return readFleetPlanJob(ctx, d, m)
}

func TestDeleteFleetPlanJob(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteFleetPlanJob(ctx, d, m)
}

// Data FleetPlan test helper functions

func TestDataFleetplan() *schema.Resource {
	return dataFleetplan()
}

func TestDataFleetplanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceFleetPlanRead(ctx, d, m)
}

func TestDataFleetplans() *schema.Resource {
	return dataFleetplans()
}

func TestDataFleetplansRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataFleetplansRead(ctx, d, m)
}

func TestDataFleetplanJob() *schema.Resource {
	return dataFleetplanJob()
}

func TestDataFleetplanJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataFleetplanJobRead(ctx, d, m)
}

func TestDataFleetplanJobs() *schema.Resource {
	return dataFleetplanJobs()
}

func TestDataFleetplanJobsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return dataFleetplanJobsRead(ctx, d, m)
}
