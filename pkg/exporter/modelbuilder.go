package exporter

import (
	"context"

	logger "github.com/dmartinol/deployment-exporter/pkg/log"
	model "github.com/dmartinol/deployment-exporter/pkg/model"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	imagesv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/rest"
)

type ModelBuilder struct {
	config *Config

	appsClient   *appsv1.AppsV1Client
	imagesClient *imagesv1.ImageV1Client
	appsV1Client *k8appsv1client.AppsV1Client
	coreClient   *corev1client.CoreV1Client

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

	builder.appsClient, err = appsv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.imagesClient, err = imagesv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.appsV1Client, err = k8appsv1client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	builder.coreClient, err = corev1client.NewForConfig(config)
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
	var namespaces *v1.NamespaceList
	var err error
	nsSelector := builder.config.NamespaceSelector()
	logger.Infof("Filtering by %s", nsSelector)
	namespaces, err = builder.coreClient.Namespaces().List(context.TODO(), metav1.ListOptions{LabelSelector: nsSelector})
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
	deployments, err := builder.appsV1Client.Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
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
	statefulSets, err := builder.appsV1Client.StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
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
	deploymentConfigs, err := builder.appsClient.DeploymentConfigs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deploymentConfig := range deploymentConfigs.Items {
		logger.Debugf("Found %s/%s", deploymentConfig.Kind, deploymentConfig.Name)
		resource := &model.DeploymentConfig{Delegate: deploymentConfig}
		builder.namespaceModel.AddResource(resource)
		builder.buildApplications(namespace, resource)
	}

	logger.Debug("=== Pods ===")
	pods, err := builder.coreClient.Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		logger.Debugf("Found %s/%s with SA %s", pod.Kind, pod.Name, pod.Spec.ServiceAccountName)
		resource := model.Pod{Delegate: pod}
		builder.namespaceModel.AddResource(resource)
	}

	return nil
}

func (builder *ModelBuilder) buildApplications(namespace string, applicationProvider model.ApplicationProvider) {
	for _, appConfig := range applicationProvider.ApplicationConfigs() {
		logger.Debugf("Loading application %s", appConfig)
		if appConfig.IsImageStream() {
			imageStream, err := builder.imagesClient.ImageStreamImages(namespace).Get(context.TODO(), appConfig.ImageStreamId(), metav1.GetOptions{})
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
