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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ListPublishingProfileXMLWithSecretsParameters are the configurable fields of a ListPublishingProfileXMLWithSecrets.
type ListPublishingProfileXMLWithSecretsParameters struct {
	ResourceGroupName string `json:"resource_group_name"`
	// +crossplane:generate:reference:type=github.com/crossplane/provider-planetscale/apis/database/v1alpha1.Database
	Database         *string         `json:"database,omitempty"`
	DatabaseRef      *xpv1.Reference `json:"databaseRef,omitempty"`
	DatabaseSelector *xpv1.Selector  `json:"databaseSelector,omitempty"`
	AppServiceName   string          `json:"app_service_name"`
}

// ListPublishingProfileXMLWithSecretsObservation are the observable fields of a ListPublishingProfileXMLWithSecrets.
type ListPublishingProfileXMLWithSecretsObservation struct {
	ProfileGotten bool `json:"profile_gotten,omitempty"`
}

// A ListPublishingProfileXMLWithSecretsSpec defines the desired state of a ListPublishingProfileXMLWithSecrets.
type ListPublishingProfileXMLWithSecretsSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ListPublishingProfileXMLWithSecretsParameters `json:"forProvider"`
}

// A ListPublishingProfileXMLWithSecretsStatus represents the observed state of a ListPublishingProfileXMLWithSecrets.
type ListPublishingProfileXMLWithSecretsStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ListPublishingProfileXMLWithSecretsObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ListPublishingProfileXMLWithSecrets is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azureextra}
type ListPublishingProfileXMLWithSecrets struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ListPublishingProfileXMLWithSecretsSpec   `json:"spec"`
	Status ListPublishingProfileXMLWithSecretsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ListPublishingProfileXMLWithSecretsList contains a list of ListPublishingProfileXMLWithSecrets
type ListPublishingProfileXMLWithSecretsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ListPublishingProfileXMLWithSecrets `json:"items"`
}

// ListPublishingProfileXMLWithSecrets type metadata.
var (
	ListPublishingProfileXMLWithSecretsKind             = reflect.TypeOf(ListPublishingProfileXMLWithSecrets{}).Name()
	ListPublishingProfileXMLWithSecretsGroupKind        = schema.GroupKind{Group: Group, Kind: ListPublishingProfileXMLWithSecretsKind}.String()
	ListPublishingProfileXMLWithSecretsKindAPIVersion   = ListPublishingProfileXMLWithSecretsKind + "." + SchemeGroupVersion.String()
	ListPublishingProfileXMLWithSecretsGroupVersionKind = SchemeGroupVersion.WithKind(ListPublishingProfileXMLWithSecretsKind)
)

func init() {
	SchemeBuilder.Register(&ListPublishingProfileXMLWithSecrets{}, &ListPublishingProfileXMLWithSecretsList{})
}
