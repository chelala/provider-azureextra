/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package listpublishingprofilexmlwithsecrets

import (
	"context"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azureextra/apis/armappservice/v1alpha1"
	apisv1alpha1 "github.com/crossplane/provider-azureextra/apis/v1alpha1"
	"github.com/crossplane/provider-azureextra/internal/features"
)

const (
	errNotListPublishingProfileXMLWithSecrets = "managed resource is not a ListPublishingProfileXMLWithSecrets custom resource"
	errTrackPCUsage                           = "cannot track ProviderConfig usage"
	errGetPC                                  = "cannot get ProviderConfig"
	errGetCreds                               = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// A NoOpService does nothing.
type WebAppsClientService struct {
	appserviceClientFactory *armappservice.ClientFactory
	webAppsClient           *armappservice.WebAppsClient
}

type AzureConfig struct {
	ClientID                       string `json:"clientId"`
	ClientSecret                   string `json:"clientSecret"`
	SubscriptionID                 string `json:"subscriptionId"`
	TenantID                       string `json:"tenantId"`
	ActiveDirectoryEndpointURL     string `json:"activeDirectoryEndpointUrl"`
	ResourceManagerEndpointURL     string `json:"resourceManagerEndpointUrl"`
	ActiveDirectoryGraphResourceID string `json:"activeDirectoryGraphResourceId"`
	SQLManagementEndpointURL       string `json:"sqlManagementEndpointUrl"`
	GalleryEndpointURL             string `json:"galleryEndpointUrl"`
	ManagementEndpointURL          string `json:"managementEndpointUrl"`
}

var (
	azureConfig AzureConfig

	newWebAppsClientService = func(creds []byte) (*WebAppsClientService, error) {
		err := json.Unmarshal(creds, &azureConfig)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling credentials")
		}
		secretCredential, err := azidentity.NewClientSecretCredential(azureConfig.TenantID, azureConfig.ClientID, azureConfig.ClientSecret, nil)
		if err != nil {
			return nil, errors.Wrap(err, "error creating secretCredential")
		}

		result := &WebAppsClientService{
			appserviceClientFactory: nil,
			webAppsClient:           nil,
		}

		result.appserviceClientFactory, err = armappservice.NewClientFactory(azureConfig.SubscriptionID, secretCredential, nil)
		if err != nil {
			return nil, errors.Wrap(err, "error creating appserviceClientFactory")
		}
		result.webAppsClient = result.appserviceClientFactory.NewWebAppsClient()

		return result, nil
	}
)

// Setup adds a controller that reconciles ListPublishingProfileXMLWithSecrets managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ListPublishingProfileXMLWithSecretsGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ListPublishingProfileXMLWithSecretsGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newWebAppsClientService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.ListPublishingProfileXMLWithSecrets{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(creds []byte) (*WebAppsClientService, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ListPublishingProfileXMLWithSecrets)
	if !ok {
		return nil, errors.New(errNotListPublishingProfileXMLWithSecrets)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(data)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *WebAppsClientService
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ListPublishingProfileXMLWithSecrets)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotListPublishingProfileXMLWithSecrets)
	}

	_, err := c.service.webAppsClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, cr.Spec.ForProvider.AppServiceName, nil)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if cr.Status.AtProvider.ProfileGotten {
		if cr.Status.AtProvider.DeletedVirtually {
			// Handle observe after is deleted virtually
			return managed.ExternalObservation{
				ResourceExists:   false,
				ResourceUpToDate: true,
			}, nil
		}

		// In this point the resource is already created and the profile is already gotten and not deleted virtually
		return managed.ExternalObservation{
			// Handle observe after created
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	} else {
		// In this point is the first time that the resource is created

		format := armappservice.PublishingProfileFormatWebDeploy
		includeIisSettings := false
		pubOpts := armappservice.CsmPublishingProfileOptions{
			Format:                           &format,
			IncludeDisasterRecoveryEndpoints: &includeIisSettings,
		}

		respPubXML, err := c.service.webAppsClient.ListPublishingProfileXMLWithSecrets(ctx, cr.Spec.ForProvider.ResourceGroupName, cr.Spec.ForProvider.AppServiceName, pubOpts, nil)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "error getting publishing profile")
		}

		data, err := io.ReadAll(respPubXML.Body)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "error reading publishing profile")
		}

		cr.Status.AtProvider.ProfileGotten = true
		cr.Status.AtProvider.DeletedVirtually = false

		cr.Status.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
			ConnectionDetails: managed.ConnectionDetails{
				"publishingProfileXML": data,
			},
		}, nil

	}
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ListPublishingProfileXMLWithSecrets)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotListPublishingProfileXMLWithSecrets)
	}

	// Won't use cr for now
	_ = cr

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ListPublishingProfileXMLWithSecrets)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotListPublishingProfileXMLWithSecrets)
	}

	// Won't use cr for now
	_ = cr

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ListPublishingProfileXMLWithSecrets)
	if !ok {
		return errors.New(errNotListPublishingProfileXMLWithSecrets)
	}

	// Mark as deleted virtually
	cr.Status.AtProvider.DeletedVirtually = true

	return nil
}
