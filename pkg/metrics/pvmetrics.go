package metrics

import (
	"github.com/kubecost/cost-model/pkg/clustercache"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	v1 "k8s.io/api/core/v1"
)

//--------------------------------------------------------------------------
//  KubePVCollector
//--------------------------------------------------------------------------

// KubePVCollector is a prometheus collector that generates PV metrics
type KubePVCollector struct {
	KubeClusterCache clustercache.ClusterCache
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (kpvcb KubePVCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("kube_persistentvolume_capacity_bytes", "The pv storage capacity in bytes", []string{}, nil)
	ch <- prometheus.NewDesc("kube_persistentvolume_status_phase", "The phase indicates if a volume is available, bound to a claim, or released by a claim.", []string{}, nil)
}

// Collect is called by the Prometheus registry when collecting metrics.
func (kpvcb KubePVCollector) Collect(ch chan<- prometheus.Metric) {
	pvs := kpvcb.KubeClusterCache.GetAllPersistentVolumes()
	for _, pv := range pvs {
		phase := pv.Status.Phase
		if phase != "" {
			phases := []struct {
				v bool
				n string
			}{
				{phase == v1.VolumePending, string(v1.VolumePending)},
				{phase == v1.VolumeAvailable, string(v1.VolumeAvailable)},
				{phase == v1.VolumeBound, string(v1.VolumeBound)},
				{phase == v1.VolumeReleased, string(v1.VolumeReleased)},
				{phase == v1.VolumeFailed, string(v1.VolumeFailed)},
			}

			for _, p := range phases {
				ch <- newKubePVStatusPhaseMetric("kube_persistentvolume_status_phase", pv.Name, p.n, boolFloat64(p.v))
			}
		}

		storage := pv.Spec.Capacity[v1.ResourceStorage]
		m := newKubePVCapacityBytesMetric("kube_persistentvolume_capacity_bytes", pv.Name, float64(storage.Value()))

		ch <- m
	}
}

//--------------------------------------------------------------------------
//  KubePVCapacityBytesMetric
//--------------------------------------------------------------------------

// KubePVCapacityBytesMetric is a prometheus.Metric
type KubePVCapacityBytesMetric struct {
	fqName string
	help   string
	pv     string
	value  float64
}

// Creates a new KubePVCapacityBytesMetric, implementation of prometheus.Metric
func newKubePVCapacityBytesMetric(fqname, pv string, value float64) KubePVCapacityBytesMetric {
	return KubePVCapacityBytesMetric{
		fqName: fqname,
		help:   "kube_persistentvolume_capacity_bytes pv storage capacity in bytes",
		pv:     pv,
		value:  value,
	}
}

// Desc returns the descriptor for the Metric. This method idempotently
// returns the same descriptor throughout the lifetime of the Metric.
func (kpcrr KubePVCapacityBytesMetric) Desc() *prometheus.Desc {
	l := prometheus.Labels{
		"persistentvolume": kpcrr.pv,
	}
	return prometheus.NewDesc(kpcrr.fqName, kpcrr.help, []string{}, l)
}

// Write encodes the Metric into a "Metric" Protocol Buffer data
// transmission object.
func (kpcrr KubePVCapacityBytesMetric) Write(m *dto.Metric) error {
	m.Gauge = &dto.Gauge{
		Value: &kpcrr.value,
	}

	m.Label = []*dto.LabelPair{
		{
			Name:  toStringPtr("persistentvolume"),
			Value: &kpcrr.pv,
		},
	}
	return nil
}

//--------------------------------------------------------------------------
//  KubePVStatusPhaseMetric
//--------------------------------------------------------------------------

// KubePVStatusPhaseMetric is a prometheus.Metric
type KubePVStatusPhaseMetric struct {
	fqName string
	help   string
	pv     string
	phase  string
	value  float64
}

// Creates a new KubePVCapacityBytesMetric, implementation of prometheus.Metric
func newKubePVStatusPhaseMetric(fqname, pv, phase string, value float64) KubePVStatusPhaseMetric {
	return KubePVStatusPhaseMetric{
		fqName: fqname,
		help:   "kube_persistentvolume_status_phase pv status phase",
		pv:     pv,
		phase:  phase,
		value:  value,
	}
}

// Desc returns the descriptor for the Metric. This method idempotently
// returns the same descriptor throughout the lifetime of the Metric.
func (kpcrr KubePVStatusPhaseMetric) Desc() *prometheus.Desc {
	l := prometheus.Labels{
		"persistentvolume": kpcrr.pv,
		"phase":            kpcrr.phase,
	}
	return prometheus.NewDesc(kpcrr.fqName, kpcrr.help, []string{}, l)
}

// Write encodes the Metric into a "Metric" Protocol Buffer data
// transmission object.
func (kpcrr KubePVStatusPhaseMetric) Write(m *dto.Metric) error {
	m.Gauge = &dto.Gauge{
		Value: &kpcrr.value,
	}

	m.Label = []*dto.LabelPair{
		{
			Name:  toStringPtr("persistentvolume"),
			Value: &kpcrr.pv,
		},
		{
			Name:  toStringPtr("phase"),
			Value: &kpcrr.phase,
		},
	}
	return nil
}
