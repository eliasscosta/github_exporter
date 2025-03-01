package exporter

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v35/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// OrgCollector collects metrics about the servers.
type OrgCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	PublicRepos       *prometheus.Desc
	PublicGists       *prometheus.Desc
	PrivateGists      *prometheus.Desc
	Followers         *prometheus.Desc
	Following         *prometheus.Desc
	Collaborators     *prometheus.Desc
	DiskUsage         *prometheus.Desc
	PrivateReposTotal *prometheus.Desc
	PrivateReposOwned *prometheus.Desc
	Created           *prometheus.Desc
	Updated           *prometheus.Desc
}

// NewOrgCollector returns a new OrgCollector.
func NewOrgCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *OrgCollector {
	if failures != nil {
		failures.WithLabelValues("org").Add(0)
	}

	labels := []string{"name"}
	return &OrgCollector{
		client:   client,
		logger:   log.With(logger, "collector", "org"),
		failures: failures,
		duration: duration,
		config:   cfg,

		PublicRepos: prometheus.NewDesc(
			"github_org_public_repos",
			"Number of public repositories from org",
			labels,
			nil,
		),
		PublicGists: prometheus.NewDesc(
			"github_org_public_gists",
			"Number of public gists from org",
			labels,
			nil,
		),
		PrivateGists: prometheus.NewDesc(
			"github_org_private_gists",
			"Number of private gists from org",
			labels,
			nil,
		),
		Followers: prometheus.NewDesc(
			"github_org_followers",
			"Number of followers for org",
			labels,
			nil,
		),
		Following: prometheus.NewDesc(
			"github_org_following",
			"Number of following other users by org",
			labels,
			nil,
		),
		Collaborators: prometheus.NewDesc(
			"github_org_collaborators",
			"Number of collaborators within org",
			labels,
			nil,
		),
		DiskUsage: prometheus.NewDesc(
			"github_org_disk_usage",
			"Used diskspace by the org",
			labels,
			nil,
		),
		PrivateReposTotal: prometheus.NewDesc(
			"github_org_private_repos_total",
			"Total amount of private repositories",
			labels,
			nil,
		),
		PrivateReposOwned: prometheus.NewDesc(
			"github_org_private_repos_owned",
			"Owned private repositories by org",
			labels,
			nil,
		),
		Created: prometheus.NewDesc(
			"github_org_create_timestamp",
			"Timestamp of the creation of org",
			labels,
			nil,
		),
		Updated: prometheus.NewDesc(
			"github_org_updated_timestamp",
			"Timestamp of the last modification of org",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *OrgCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.PublicRepos,
		c.PublicGists,
		c.PrivateGists,
		c.Followers,
		c.Following,
		c.Collaborators,
		c.DiskUsage,
		c.PrivateReposTotal,
		c.PrivateReposOwned,
		c.Created,
		c.Updated,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *OrgCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.PublicRepos
	ch <- c.PublicGists
	ch <- c.PrivateGists
	ch <- c.Followers
	ch <- c.Following
	ch <- c.Collaborators
	ch <- c.DiskUsage
	ch <- c.PrivateReposTotal
	ch <- c.PrivateReposOwned
	ch <- c.Created
	ch <- c.Updated
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *OrgCollector) Collect(ch chan<- prometheus.Metric) {
	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		now := time.Now()
		record, _, err := c.client.Organizations.Get(ctx, name)
		c.duration.WithLabelValues("org").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("org").Inc()
			continue
		}

		labels := []string{
			name,
		}

		if record.PublicRepos != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PublicRepos,
				prometheus.GaugeValue,
				float64(*record.PublicRepos),
				labels...,
			)
		}

		if record.PublicGists != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PublicGists,
				prometheus.GaugeValue,
				float64(*record.PublicGists),
				labels...,
			)
		}

		if record.PrivateGists != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PrivateGists,
				prometheus.GaugeValue,
				float64(*record.PrivateGists),
				labels...,
			)
		}

		if record.Followers != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Followers,
				prometheus.GaugeValue,
				float64(*record.Followers),
				labels...,
			)
		}

		if record.Following != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Following,
				prometheus.GaugeValue,
				float64(*record.Following),
				labels...,
			)
		}

		if record.Collaborators != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Collaborators,
				prometheus.GaugeValue,
				float64(*record.Collaborators),
				labels...,
			)
		}

		if record.DiskUsage != nil {
			ch <- prometheus.MustNewConstMetric(
				c.DiskUsage,
				prometheus.GaugeValue,
				float64(*record.DiskUsage),
				labels...,
			)
		}

		if record.TotalPrivateRepos != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PrivateReposTotal,
				prometheus.GaugeValue,
				float64(*record.TotalPrivateRepos),
				labels...,
			)
		}

		if record.OwnedPrivateRepos != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PrivateReposOwned,
				prometheus.GaugeValue,
				float64(*record.OwnedPrivateRepos),
				labels...,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Created,
			prometheus.GaugeValue,
			float64(record.CreatedAt.Unix()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Updated,
			prometheus.GaugeValue,
			float64(record.UpdatedAt.Unix()),
			labels...,
		)
	}
}
