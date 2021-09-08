/*


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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BuckySpec defines the desired state of Bucky
type BuckySpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1

	SeleniumNodeNumber int `json:"seleniumNodeNumber"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1

	NodeInstanceNumber int `json:"nodeInstanceNumber"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Formate:=string

	BuckyCoreImage string `json:"buckyCoreImage"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Formate:=string

	BuckyCommand string `json:"buckyCommand"`
}

// BuckyStatus defines the observed state of Bucky
type BuckyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Bucky is the Schema for the buckys API
type Bucky struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuckySpec   `json:"spec,omitempty"`
	Status BuckyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BuckyList contains a list of Bucky
type BuckyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucky `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucky{}, &BuckyList{})
}
