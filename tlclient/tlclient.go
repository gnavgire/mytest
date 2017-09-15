/*
Copyright 2016 Iguazio Systems Ltd.

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
package tlclient

import (

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
        "k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
        CITTLPlural string = "trusttabs"
        CITTLGroup  string = "cit.intel.com"
        CITTLVersion string = "v1beta1"
)

func CitTLClient(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *cittlclient {
	return &cittlclient{cl: cl, ns: namespace, plural: CITTLPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type cittlclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

// Definition of our CRD Example class
type Trusttab struct {
        meta_v1.TypeMeta   `json:",inline"`
        meta_v1.ObjectMeta `json:"metadata"`
        Spec               trusttabspec   `json:"spec"`
        Status             trusttabstatus `json:"status,omitempty"`
}
type trusttabspec struct {
        hostlist string `json:"hostlist"`
        trusttabs string   `json:"trusttabs"`
}

type trusttabstatus struct {
        State   string `json:"state,omitempty"`
        Message string `json:"message,omitempty"`
}

type TrustTabList struct {
        meta_v1.TypeMeta `json:",inline"`
        meta_v1.ListMeta `json:"metadata"`
        Items            []Trusttab `json:"items"`
}

// Create a  Rest client with the new CRD Schema
var SchemeGroupVersion = schema.GroupVersion{Group: CITTLGroup, Version: CITTLVersion}

func addKnownTypes(scheme *runtime.Scheme) error {
        scheme.AddKnownTypes(SchemeGroupVersion,
                &Trusttab{},
                &TrustTabList{},
        )
        meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
        return nil
}

func NewClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
        scheme := runtime.NewScheme()
        SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
        if err := SchemeBuilder.AddToScheme(scheme); err != nil {
                return nil, nil, err
        }
        config := *cfg
        config.GroupVersion = &SchemeGroupVersion
        config.APIPath = "/apis"
        config.ContentType = runtime.ContentTypeJSON
        config.NegotiatedSerializer = serializer.DirectCodecFactory{
                CodecFactory: serializer.NewCodecFactory(scheme)}

        client, err := rest.RESTClientFor(&config)
        if err != nil {
                return nil, nil, err
        }
        return client, scheme, nil
}

/*

func (f *cittlclient) Create(obj *Trusttab) (*Trusttab, error) {
	var result Trusttab
	err := f.cl.Post().
		Namespace(f.ns).Resource(f.plural).
		Body(obj).Do().Into(&result)
	return &result, err
}

func (f *cittlclient) Update(obj *Trusttab) (*Trusttab, error) {
	var result Trusttab
	err := f.cl.Put().
		Namespace(f.ns).Resource(f.plural).
		Body(obj).Do().Into(&result)
	return &result, err
}

func (f *cittlclient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.cl.Delete().
		Namespace(f.ns).Resource(f.plural).
		Name(name).Body(options).Do().
		Error()
}


func (f *cittlclient) Get(name string) (*Trusttab, error) {
	var result Trusttab
	err := f.cl.Get().
		Namespace(f.ns).Resource(f.plural).
		Name(name).Do().Into(&result)
	return &result, err
}

func (f *cittlclient) List(opts meta_v1.ListOptions) (*TrustTabList, error) {
	var result TrustTabList
	err := f.cl.Get().
		Namespace(f.ns).Resource(f.plural).
		VersionedParams(&opts, f.codec).
		Do().Into(&result)
	return &result, err
}


*/

// Create a new List watch for our TPR
func (f *cittlclient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}

