package job

import (
	"context"

	scrapeconfig "github.com/SchwarzIT/community-stackit-go-client/pkg/services/argus/v1.0/generated/scrape-config"
	"github.com/SchwarzIT/terraform-provider-stackit/stackit/internal/common"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	default_metrics_path                = "/metrics"
	default_scheme                      = "https"
	default_scrape_interval             = "5m"
	default_scrape_timeout              = "2m"
	default_saml2_enable_url_parameters = true
)

func (j *Job) setDefaults(job *scrapeconfig.CreateJSONBody) {
	if job == nil {
		return
	}
	if j.MetricsPath.IsNull() || j.MetricsPath.IsUnknown() {
		s := default_metrics_path
		job.MetricsPath = &s
	}
	if j.Scheme.IsNull() || j.Scheme.IsUnknown() {
		job.Scheme = scrapeconfig.CreateJSONBodyScheme(default_scheme)
	}
	if j.ScrapeInterval.IsNull() || j.ScrapeInterval.IsUnknown() {
		job.ScrapeInterval = default_scrape_interval
	}
	if j.ScrapeTimeout.IsNull() || j.ScrapeTimeout.IsUnknown() {
		job.ScrapeTimeout = default_scrape_timeout
	}
}

func (j *Job) setDefaultsUpdate(job *scrapeconfig.UpdateJSONBody) {
	if job == nil {
		return
	}
	if j.MetricsPath.IsNull() || j.MetricsPath.IsUnknown() {
		job.MetricsPath = default_metrics_path
	}
	if j.Scheme.IsNull() || j.Scheme.IsUnknown() {
		job.Scheme = scrapeconfig.UpdateJSONBodyScheme(default_scheme)
	}
	if j.ScrapeInterval.IsNull() || j.ScrapeInterval.IsUnknown() {
		job.ScrapeInterval = default_scrape_interval
	}
	if j.ScrapeTimeout.IsNull() || j.ScrapeTimeout.IsUnknown() {
		job.ScrapeTimeout = default_scrape_timeout
	}
}

func (j *Job) ToClientJob() scrapeconfig.CreateJSONBody {
	mp := j.MetricsPath.ValueString()
	job := scrapeconfig.CreateJSONBody{
		JobName:        j.Name.ValueString(),
		Scheme:         scrapeconfig.CreateJSONBodyScheme(j.Scheme.ValueString()),
		MetricsPath:    &mp,
		ScrapeInterval: j.ScrapeInterval.ValueString(),
		ScrapeTimeout:  j.ScrapeTimeout.ValueString(),
	}

	j.setDefaults(&job)

	if j.SAML2 != nil && !j.SAML2.EnableURLParameters.ValueBool() {
		if job.Params == nil {
			job.Params = &map[string]interface{}{}
		}
		p := *job.Params
		p["saml2"] = []string{"disabled"}
		job.Params = &p
	}

	if j.BasicAuth != nil {
		if job.BasicAuth == nil {
			u := j.BasicAuth.Username.ValueString()
			p := j.BasicAuth.Password.ValueString()
			job.BasicAuth = &struct {
				Password *string `json:"password,omitempty"`
				Username *string `json:"username,omitempty"`
			}{
				Username: &u,
				Password: &p,
			}
		}
	}

	t := make([]struct {
		Labels  *map[string]interface{} `json:"labels,omitempty"`
		Targets []string                `json:"targets"`
	}, len(j.Targets))
	for i, target := range j.Targets {
		ti := struct {
			Labels  *map[string]interface{} `json:"labels,omitempty"`
			Targets []string                `json:"targets"`
		}{}
		ti.Targets = make([]string, len(target.URLs))
		for k, v := range target.URLs {
			ti.Targets[k] = v.ValueString()
		}

		ls := map[string]interface{}{}
		for k, v := range target.Labels.Elements() {
			ls[k], _ = common.ToString(context.TODO(), v)
		}
		ti.Labels = &ls
		t[i] = ti
	}
	job.StaticConfigs = t
	return job
}

func (j *Job) ToClientUpdateJob() scrapeconfig.UpdateJSONBody {
	job := scrapeconfig.UpdateJSONBody{
		Scheme:         scrapeconfig.UpdateJSONBodyScheme(j.Scheme.ValueString()),
		MetricsPath:    j.MetricsPath.ValueString(),
		ScrapeInterval: j.ScrapeInterval.ValueString(),
		ScrapeTimeout:  j.ScrapeTimeout.ValueString(),
	}

	j.setDefaultsUpdate(&job)

	if j.SAML2 != nil && !j.SAML2.EnableURLParameters.ValueBool() {
		if job.Params == nil {
			job.Params = &map[string]interface{}{}
		}
		p := *job.Params
		p["saml2"] = []string{"disabled"}
		job.Params = &p
	}

	if j.BasicAuth != nil {
		if job.BasicAuth == nil {
			u := j.BasicAuth.Username.ValueString()
			p := j.BasicAuth.Password.ValueString()
			job.BasicAuth = &struct {
				Password *string `json:"password,omitempty"`
				Username *string `json:"username,omitempty"`
			}{
				Username: &u,
				Password: &p,
			}
		}
	}

	t := make([]struct {
		Labels  *map[string]interface{} `json:"labels,omitempty"`
		Targets []string                `json:"targets"`
	}, len(j.Targets))
	for i, target := range j.Targets {
		ti := struct {
			Labels  *map[string]interface{} `json:"labels,omitempty"`
			Targets []string                `json:"targets"`
		}{}
		ti.Targets = make([]string, len(target.URLs))
		for k, v := range target.URLs {
			ti.Targets[k] = v.ValueString()
		}

		ls := map[string]interface{}{}
		for k, v := range target.Labels.Elements() {
			ls[k], _ = common.ToString(context.TODO(), v)
		}
		ti.Labels = &ls
		t[i] = ti
	}
	job.StaticConfigs = t
	return job
}

func (j *Job) FromClientJob(cj scrapeconfig.Job) {
	j.ID = types.StringValue(cj.JobName)
	j.Name = types.StringValue(cj.JobName)
	if cj.MetricsPath != nil {
		j.MetricsPath = types.StringValue(*cj.MetricsPath)
	}
	if cj.Scheme != nil {
		j.Scheme = types.StringValue(string(*cj.Scheme))
	}
	j.ScrapeInterval = types.StringValue(cj.ScrapeInterval)
	j.ScrapeTimeout = types.StringValue(cj.ScrapeTimeout)
	j.handleSAML2(cj)
	j.handleBasicAuth(cj)
	j.handleTargets(cj)
}

func (j *Job) handleBasicAuth(cj scrapeconfig.Job) {
	if cj.BasicAuth == nil {
		j.BasicAuth = nil
		return
	}
	j.BasicAuth = &BasicAuth{
		Username: types.StringValue(cj.BasicAuth.Username),
		Password: types.StringValue(cj.BasicAuth.Password),
	}
}

func (j *Job) handleSAML2(cj scrapeconfig.Job) {
	if cj.Params == nil && j.SAML2 == nil {
		return
	}

	if j.SAML2 == nil {
		j.SAML2 = &SAML2{}
	}

	flag := true
	p := *cj.Params
	if v, ok := p["saml2"]; ok {
		if len(v) == 1 && v[0] == "disabled" {
			flag = false
		}
	}

	j.SAML2 = &SAML2{
		EnableURLParameters: types.BoolValue(flag),
	}
}

func (j *Job) handleTargets(cj scrapeconfig.Job) {
	newTargets := []Target{}
	for i, sc := range cj.StaticConfigs {
		nt := Target{
			URLs: []types.String{},
		}
		for _, v := range sc.Targets {
			nt.URLs = append(nt.URLs, types.StringValue(v))
		}

		if len(j.Targets) > i && j.Targets[i].Labels.IsNull() {
			nt.Labels = j.Targets[i].Labels
		} else {
			nt.Labels, _ = types.MapValueFrom(context.TODO(), types.StringType, sc.Labels)
		}
		newTargets = append(newTargets, nt)
	}
	j.Targets = newTargets
}
