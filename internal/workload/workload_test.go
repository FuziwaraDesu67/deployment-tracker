package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
)

func TestHasSupportedOwner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
	}{
		{
			name: "pod owned by ReplicaSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "test-rs-abc123",
					}},
				},
			},
			expected: true,
		},
		{
			name: "pod owned by Job",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "Job",
						Name: "test-job",
					}},
				},
			},
			expected: true,
		},
		{
			name: "pod with no owner",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: false,
		},
		{
			name: "pod owned by DaemonSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "DaemonSet",
						Name: "test-ds",
					}},
				},
			},
			expected: true,
		},
		{
			name: "pod owned by StatefulSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "test-ss",
					}},
				},
			},
			expected: true,
		},
		{
			name: "pod owned by ReplicationController",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicationController",
						Name: "test-rc",
					}},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := HasSupportedOwner(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetJobOwnerName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "pod owned by Job",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "Job",
						Name: "my-job",
					}},
				},
			},
			expected: "my-job",
		},
		{
			name: "pod not owned by Job",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "my-rs-abc123",
					}},
				},
			},
			expected: "",
		},
		{
			name: "pod with no owner",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetJobOwnerName(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDeploymentName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "pod owned by ReplicaSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "my-deployment-abc123",
					}},
				},
			},
			expected: "my-deployment",
		},
		{
			name: "pod not owned by ReplicaSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "DaemonSet",
						Name: "my-ds",
					}},
				},
			},
			expected: "",
		},
		{
			name: "pod with no owner",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetDeploymentName(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDaemonSetName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "pod owned by DaemonSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "DaemonSet",
						Name: "my-daemonset",
					}},
				},
			},
			expected: "my-daemonset",
		},
		{
			name: "pod not owned by DaemonSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "my-rs-abc123",
					}},
				},
			},
			expected: "",
		},
		{
			name: "pod with no owner",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetDaemonSetName(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStatefulSetName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
	}{
		{
			name: "pod owned by StatefulSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "my-statefulset",
					}},
				},
			},
			expected: "my-statefulset",
		},
		{
			name: "pod not owned by StatefulSet",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "my-rs-abc123",
					}},
				},
			},
			expected: "",
		},
		{
			name: "pod with no owner",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetStatefulSetName(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTerminalPhase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		phase    corev1.PodPhase
		expected bool
	}{
		{"Succeeded", corev1.PodSucceeded, true},
		{"Failed", corev1.PodFailed, true},
		{"Running", corev1.PodRunning, false},
		{"Pending", corev1.PodPending, false},
		{"Unknown", corev1.PodUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pod := &corev1.Pod{
				Status: corev1.PodStatus{Phase: tt.phase},
			}
			assert.Equal(t, tt.expected, IsTerminalPhase(pod))
		})
	}
}

func TestIsNumeric(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"digits", "28485120", true},
		{"zero", "0", true},
		{"long digits", "123456789", true},
		{"empty string", "", false},
		{"letters", "abc", false},
		{"mixed", "123abc", false},
		{"hyphenated", "12-34", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, isNumeric(tt.input))
		})
	}
}

func TestResolve_Deployment(t *testing.T) {
	t.Parallel()
	r := &Resolver{}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "ReplicaSet",
				Name: "my-deployment-abc123",
			}},
		},
	}
	ref := r.Resolve(pod)
	assert.Equal(t, "my-deployment", ref.Name)
	assert.Equal(t, "Deployment", ref.Kind)
}

func TestResolve_DaemonSet(t *testing.T) {
	t.Parallel()
	r := &Resolver{}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "DaemonSet",
				Name: "my-daemonset",
			}},
		},
	}
	ref := r.Resolve(pod)
	assert.Equal(t, "my-daemonset", ref.Name)
	assert.Equal(t, "DaemonSet", ref.Kind)
}

func TestResolve_StatefulSet(t *testing.T) {
	t.Parallel()
	r := &Resolver{}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "StatefulSet",
				Name: "my-statefulset",
			}},
		},
	}
	ref := r.Resolve(pod)
	assert.Equal(t, "my-statefulset", ref.Name)
	assert.Equal(t, "StatefulSet", ref.Kind)
}

func TestResolve_StandaloneJob(t *testing.T) {
	t.Parallel()
	// With nil listers, resolveJobWorkload falls back to standalone Job
	r := &Resolver{}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "Job",
				Name: "my-standalone-job",
			}},
		},
	}
	ref := r.Resolve(pod)
	assert.Equal(t, "my-standalone-job", ref.Name)
	assert.Equal(t, "Job", ref.Kind)
}

func TestResolve_NoOwner(t *testing.T) {
	t.Parallel()
	r := &Resolver{}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{},
	}
	ref := r.Resolve(pod)
	assert.Empty(t, ref.Name)
	assert.Empty(t, ref.Kind)
}

func TestIsActive_Deployment(t *testing.T) {
	t.Parallel()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	activeDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "active-dep",
			Namespace: "default",
		},
	}
	now := metav1.Now()
	terminatingDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "terminating-dep",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
	}
	require.NoError(t, indexer.Add(activeDep))
	require.NoError(t, indexer.Add(terminatingDep))

	r := NewResolver(appslisters.NewDeploymentLister(indexer), nil, nil, nil, nil)
	assert.True(t, r.IsActive("default", Identity{Name: "active-dep", Kind: "Deployment"}))
	assert.False(t, r.IsActive("default", Identity{Name: "terminating-dep", Kind: "Deployment"}))
	assert.False(t, r.IsActive("default", Identity{Name: "nonexistent", Kind: "Deployment"}))
}

func TestIsActive_DaemonSet(t *testing.T) {
	t.Parallel()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	activeDS := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "active-ds",
			Namespace: "default",
		},
	}
	now := metav1.Now()
	terminatingDS := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "terminating-ds",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
	}
	require.NoError(t, indexer.Add(activeDS))
	require.NoError(t, indexer.Add(terminatingDS))

	r := NewResolver(nil, appslisters.NewDaemonSetLister(indexer), nil, nil, nil)
	assert.True(t, r.IsActive("default", Identity{Name: "active-ds", Kind: "DaemonSet"}))
	assert.False(t, r.IsActive("default", Identity{Name: "terminating-ds", Kind: "DaemonSet"}))
	assert.False(t, r.IsActive("default", Identity{Name: "nonexistent", Kind: "DaemonSet"}))
}

func TestIsActive_StatefulSet(t *testing.T) {
	t.Parallel()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	activeSS := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "active-ss",
			Namespace: "default",
		},
	}
	now := metav1.Now()
	terminatingSS := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "terminating-ss",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
	}
	require.NoError(t, indexer.Add(activeSS))
	require.NoError(t, indexer.Add(terminatingSS))

	r := NewResolver(nil, nil, appslisters.NewStatefulSetLister(indexer), nil, nil)
	assert.True(t, r.IsActive("default", Identity{Name: "active-ss", Kind: "StatefulSet"}))
	assert.False(t, r.IsActive("default", Identity{Name: "terminating-ss", Kind: "StatefulSet"}))
	assert.False(t, r.IsActive("default", Identity{Name: "nonexistent", Kind: "StatefulSet"}))
}

func TestIsActive_Job(t *testing.T) {
	t.Parallel()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	activeJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "active-job",
			Namespace: "default",
		},
	}
	now := metav1.Now()
	terminatingJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "terminating-job",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
	}
	require.NoError(t, indexer.Add(activeJob))
	require.NoError(t, indexer.Add(terminatingJob))

	r := NewResolver(nil, nil, nil, batchlisters.NewJobLister(indexer), nil)
	assert.True(t, r.IsActive("default", Identity{Name: "active-job", Kind: "Job"}))
	assert.False(t, r.IsActive("default", Identity{Name: "terminating-job", Kind: "Job"}))
	assert.False(t, r.IsActive("default", Identity{Name: "nonexistent", Kind: "Job"}))
}

func TestIsActive_CronJob(t *testing.T) {
	t.Parallel()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	activeCronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "active-cronjob",
			Namespace: "default",
		},
	}
	now := metav1.Now()
	terminatingCronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "terminating-cronjob",
			Namespace:         "default",
			DeletionTimestamp: &now,
		},
	}
	require.NoError(t, indexer.Add(activeCronJob))
	require.NoError(t, indexer.Add(terminatingCronJob))

	r := NewResolver(nil, nil, nil, nil, batchlisters.NewCronJobLister(indexer))
	assert.True(t, r.IsActive("default", Identity{Name: "active-cronjob", Kind: "CronJob"}))
	assert.False(t, r.IsActive("default", Identity{Name: "terminating-cronjob", Kind: "CronJob"}))
	assert.False(t, r.IsActive("default", Identity{Name: "nonexistent", Kind: "CronJob"}))
}

func TestIsActive_NilListersAndUnknownKind(t *testing.T) {
	t.Parallel()
	r := &Resolver{}
	assert.False(t, r.IsActive("default", Identity{Name: "test", Kind: "Deployment"}))
	assert.False(t, r.IsActive("default", Identity{Name: "test", Kind: "DaemonSet"}))
	assert.False(t, r.IsActive("default", Identity{Name: "test", Kind: "StatefulSet"}))
	assert.False(t, r.IsActive("default", Identity{Name: "test", Kind: "Job"}))
	assert.False(t, r.IsActive("default", Identity{Name: "test", Kind: "CronJob"}))
	assert.False(t, r.IsActive("default", Identity{Name: "test", Kind: "UnknownKind"}))
}
