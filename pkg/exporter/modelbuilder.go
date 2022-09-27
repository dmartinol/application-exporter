package exporter

import (
	"context"

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
	config *Config

	clientAppsV1       *clientAppsV1.AppsV1Client
	clientImagesV1     *clientImagesV1.ImageV1Client
	k8sAppsClientV1    *k8sClientAppsV1.AppsV1Client
	k8sBatchClientV1   *k8sClientBatchV1.BatchV1Client
	k8sCoreClientV1    *k8sClientCoreV1.CoreV1Client
	k8sMetricsClientV1 *k8sClientMetrics.Clientset

	topologyModel  *model.TopologyModel
	namespaceModel *model.NamespaceModel
}

func NewModelBuilder(config *Config) *ModelBuilder {
	builder := ModelBuilder{config: config}
	builder.topologyModel = model.NewTopologyModel()
	return &builder
}

func (builder *ModelBuilder) BuildForKubeConfig(config *rest.Config) (*model.TopologyModel, error) {
	var err error

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
	var namespaces *k8sCoreV1.NamespaceList
	var err error
	nsSelector := builder.config.NamespaceSelector()
	logger.Infof("Filtering by %s", nsSelector)
	namespaces, err = builder.k8sCoreClientV1.Namespaces().List(context.TODO(), k8sMetaV1.ListOptions{LabelSelector: nsSelector})
	if err != nil {
		logger.Warnf("Cannot list namespaces by selector %s: %s", nsSelector, err)
		return err
	}
	for _, namespace := range namespaces.Items {
		err := builder.buildNamespace(namespace.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (builder *ModelBuilder) buildNamespace(namespace string) error {
	builder.namespaceModel = builder.topologyModel.AddNamespace(namespace)

	logger.Infof("Running on NS %s", namespace)
	logger.Debug("=== Deployments ===")
	deployments, err := builder.k8sAppsClientV1.Deployments(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deployment := range deployments.Items {
		logger.Debugf("Found %s/%s", deployment.Kind, deployment.Name)
		resource := &model.Deployment{Delegate: deployment}
		builder.namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debug("=== StatefulSets ===")
	statefulSets, err := builder.k8sAppsClientV1.StatefulSets(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, statefulSet := range statefulSets.Items {
		logger.Debugf("Found %s/%s", statefulSet.Kind, statefulSet.Name)
		resource := model.StatefulSet{Delegate: statefulSet}
		builder.namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debug("=== DeploymentConfigs ===")
	deploymentConfigs, err := builder.clientAppsV1.DeploymentConfigs(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		logger.Debugf("Found %s/%s", deploymentConfig.Kind, deploymentConfig.Name)
		resource := &model.DeploymentConfig{Delegate: deploymentConfig}
		builder.namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debug("=== CronJobs ===")
	cronJobs, err := builder.k8sBatchClientV1.CronJobs(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cronJob := range cronJobs.Items {
		logger.Debugf("Found %s/%s", cronJob.Kind, cronJob.Name)
		resource := &model.CronJob{Delegate: cronJob}
		builder.namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debug("=== DaemonSets ===")
	demonSets, err := builder.k8sAppsClientV1.DaemonSets(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		return err
	}
	for _, demonSet := range demonSets.Items {
		logger.Debugf("Found %s/%s", demonSet.Kind, demonSet.Name)
		resource := &model.DaemonSet{Delegate: demonSet}
		builder.namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debug("=== Pods ===")
	pods, err := builder.k8sCoreClientV1.Pods(namespace).List(context.TODO(), k8sMetaV1.ListOptions{})
	if err != nil {
		return err
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
		builder.namespaceModel.AddResource(resource)
	}

	return nil
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
