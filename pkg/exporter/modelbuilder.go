package exporter

import (
	"context"
	"sync"
	"time"

	"github.com/dmartinol/application-exporter/pkg/config"
	logger "github.com/dmartinol/application-exporter/pkg/log"
	model "github.com/dmartinol/application-exporter/pkg/model"
	clientAppsV1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	clientImagesV1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	k8sCoreV1 "k8s.io/api/core/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClientAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	k8sClientBatchV1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	k8sClientCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sClientMetrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"k8s.io/client-go/rest"
)

type ModelBuilder struct {
	config       *config.Config
	runnerConfig *config.RunnerConfig

	clientAppsV1       *clientAppsV1.AppsV1Client
	clientImagesV1     *clientImagesV1.ImageV1Client
	k8sAppsClientV1    *k8sClientAppsV1.AppsV1Client
	k8sBatchClientV1   *k8sClientBatchV1.BatchV1Client
	k8sCoreClientV1    *k8sClientCoreV1.CoreV1Client
	k8sMetricsClientV1 *k8sClientMetrics.Clientset

	topologyModel *model.TopologyModel
}

func NewModelBuilder(config *config.Config, runnerConfig *config.RunnerConfig) *ModelBuilder {
	builder := ModelBuilder{config: config, runnerConfig: runnerConfig}
	builder.topologyModel = model.NewTopologyModel()
	return &builder
}

func (builder *ModelBuilder) BuildForKubeConfig(config *rest.Config) (*model.TopologyModel, error) {
	var err error
	config.Burst = builder.config.Burst()

	builder.clientAppsV1, err = clientAppsV1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.clientImagesV1, err = clientImagesV1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.k8sAppsClientV1, err = k8sClientAppsV1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.k8sBatchClientV1, err = k8sClientBatchV1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.k8sCoreClientV1, err = k8sClientCoreV1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.k8sMetricsClientV1, err = k8sClientMetrics.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	err = builder.buildCluster()
	if err != nil {
		return nil, err
	}

	return builder.topologyModel, nil
}

func (builder *ModelBuilder) buildCluster() error {
	logger.Infof("Starting data collection for:\n%s\n%s", builder.config, builder.runnerConfig)
	startAt := time.Now()
	var namespaces *k8sCoreV1.NamespaceList
	var err error
	nsSelector := builder.runnerConfig.NamespaceSelector()
	logger.Infof("Filtering by %s", nsSelector)
	namespaces, err = builder.k8sCoreClientV1.Namespaces().List(context.TODO(), k8sMetaV1.ListOptions{LabelSelector: nsSelector})
	if err != nil {
		logger.Warnf("Cannot list namespaces by selector %s: %s", nsSelector, err)
		return err
	}

	wg := new(sync.WaitGroup)

	nsErr := make(chan error, len(namespaces.Items))
	for _, namespace := range namespaces.Items {
		wg.Add(1)
		go builder.buildNamespace(wg, namespace.Name, nsErr)
	}
	wg.Wait()
	close(nsErr)
	var open bool
	if err, open = <-nsErr; open {
		return err
	}

	duration := time.Since(startAt)
	logger.Infof("Data collection completed in %s (max burst is %d)", duration, builder.config.Burst())

	return nil
}

func (builder *ModelBuilder) buildNamespace(wg *sync.WaitGroup, namespace string, nsErr chan error) {
	defer wg.Done()
	namespaceModel := builder.topologyModel.AddNamespace(namespace)

	logger.Infof("Running on NS %s", namespace)
	logger.Debugf("=== %s Deployments ===", namespace)
	deployments, err := builder.k8sAppsClientV1.Deployments(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		nsErr <- err
		return
	}
	for _, deployment := range deployments.Items {
		logger.Debugf("Found %s/%s", deployment.Kind, deployment.Name)
		resource := &model.Deployment{Delegate: deployment}
		namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debugf("=== %s StatefulSets ===", namespace)
	statefulSets, err := builder.k8sAppsClientV1.StatefulSets(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		nsErr <- err
		return
	}
	for _, statefulSet := range statefulSets.Items {
		logger.Debugf("Found %s/%s", statefulSet.Kind, statefulSet.Name)
		resource := model.StatefulSet{Delegate: statefulSet}
		namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debugf("=== %s DeploymentConfigs ===", namespace)
	deploymentConfigs, err := builder.clientAppsV1.DeploymentConfigs(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		nsErr <- err
		return
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		logger.Debugf("Found %s/%s", deploymentConfig.Kind, deploymentConfig.Name)
		resource := &model.DeploymentConfig{Delegate: deploymentConfig}
		namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debugf("=== %s CronJobs ===", namespace)
	cronJobs, err := builder.k8sBatchClientV1.CronJobs(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		nsErr <- err
		return
	}
	for _, cronJob := range cronJobs.Items {
		logger.Debugf("Found %s/%s", cronJob.Kind, cronJob.Name)
		resource := &model.CronJob{Delegate: cronJob}
		namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debugf("=== %s DaemonSets ===", namespace)
	demonSets, err := builder.k8sAppsClientV1.DaemonSets(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		nsErr <- err
		return
	}
	for _, demonSet := range demonSets.Items {
		logger.Debugf("Found %s/%s", demonSet.Kind, demonSet.Name)
		resource := &model.DaemonSet{Delegate: demonSet}
		namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debugf("=== %s Pods ===", namespace)
	pods, err := builder.k8sCoreClientV1.Pods(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		nsErr <- err
		return
	}
	for _, pod := range pods.Items {
		logger.Debugf("Found %s/%s with SA %s", pod.Kind, pod.Name, pod.Spec.ServiceAccountName)
		resource := model.Pod{Delegate: pod}
		if builder.config.WithResources() && resource.IsRunning() {
			podMetrics, err := builder.k8sMetricsClientV1.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), pod.Name, k8sMetaV1.GetOptions{})
			if err != nil {
				logger.Warnf("No metrics for Pod %s: %s", pod.Name, err)
			} else {
				resource.SetMetrics(podMetrics)
			}
		}
		namespaceModel.AddResource(resource)
	}

	logger.Infof("Completed NS %s", namespace)
}

func (builder *ModelBuilder) buildApplications(namespace string, applicationProvider model.ApplicationProvider) {
	for _, appConfig := range applicationProvider.ApplicationConfigs() {
		logger.Debugf("Loading application %s", appConfig)
		if appConfig.IsImageStream() {
			imageStream, err := builder.clientImagesV1.ImageStreamImages(namespace).Get(context.TODO(), appConfig.ImageStreamId(), k8sMetaV1.GetOptions{})
			if err != nil {
				logger.Warnf("Cannot load image for %s: %s", appConfig.ImageName, err)
			} else {
				logger.Debugf("Found image %s", imageStream.Image.Name)
				builder.topologyModel.AddImage(appConfig.ImageName, model.NewImageByStream(appConfig.ImageName, imageStream.Image))
			}
		} else {
			builder.topologyModel.AddImage(appConfig.ImageName, model.NewImageByRegistry(appConfig.ImageName))
		}
	}
}
