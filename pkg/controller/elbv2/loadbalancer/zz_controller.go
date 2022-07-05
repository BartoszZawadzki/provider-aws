/*
Copyright 2021 The Crossplane Authors.

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

// Code generated by ack-generate. DO NOT EDIT.

package loadbalancer

import (
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/elbv2"
	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

const (
	errUnexpectedObject = "managed resource is not an LoadBalancer resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create LoadBalancer in AWS"
	errUpdate        = "cannot update LoadBalancer in AWS"
	errDescribe      = "failed to describe LoadBalancer"
	errDelete        = "failed to delete LoadBalancer"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.LoadBalancer)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	sess, err := awsclient.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.LoadBalancer)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateDescribeLoadBalancersInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.DescribeLoadBalancersWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	resp = e.filterList(cr, resp)
	if len(resp.LoadBalancers) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateLoadBalancer(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)

	upToDate, err := e.isUpToDate(cr, resp)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
	}
	return e.postObserve(ctx, cr, resp, managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil)
}

func (e *external) Create(ctx context.Context, mg cpresource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.LoadBalancer)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateLoadBalancerInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateLoadBalancerWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	found := false
	for _, elem := range resp.LoadBalancers {
		if elem.AvailabilityZones != nil {
			f0 := []*svcapitypes.AvailabilityZone{}
			for _, f0iter := range elem.AvailabilityZones {
				f0elem := &svcapitypes.AvailabilityZone{}
				if f0iter.LoadBalancerAddresses != nil {
					f0elemf0 := []*svcapitypes.LoadBalancerAddress{}
					for _, f0elemf0iter := range f0iter.LoadBalancerAddresses {
						f0elemf0elem := &svcapitypes.LoadBalancerAddress{}
						if f0elemf0iter.AllocationId != nil {
							f0elemf0elem.AllocationID = f0elemf0iter.AllocationId
						}
						if f0elemf0iter.IPv6Address != nil {
							f0elemf0elem.IPv6Address = f0elemf0iter.IPv6Address
						}
						if f0elemf0iter.IpAddress != nil {
							f0elemf0elem.IPAddress = f0elemf0iter.IpAddress
						}
						if f0elemf0iter.PrivateIPv4Address != nil {
							f0elemf0elem.PrivateIPv4Address = f0elemf0iter.PrivateIPv4Address
						}
						f0elemf0 = append(f0elemf0, f0elemf0elem)
					}
					f0elem.LoadBalancerAddresses = f0elemf0
				}
				if f0iter.OutpostId != nil {
					f0elem.OutpostID = f0iter.OutpostId
				}
				if f0iter.SubnetId != nil {
					f0elem.SubnetID = f0iter.SubnetId
				}
				if f0iter.ZoneName != nil {
					f0elem.ZoneName = f0iter.ZoneName
				}
				f0 = append(f0, f0elem)
			}
			cr.Status.AtProvider.AvailabilityZones = f0
		} else {
			cr.Status.AtProvider.AvailabilityZones = nil
		}
		if elem.CanonicalHostedZoneId != nil {
			cr.Status.AtProvider.CanonicalHostedZoneID = elem.CanonicalHostedZoneId
		} else {
			cr.Status.AtProvider.CanonicalHostedZoneID = nil
		}
		if elem.CreatedTime != nil {
			cr.Status.AtProvider.CreatedTime = &metav1.Time{*elem.CreatedTime}
		} else {
			cr.Status.AtProvider.CreatedTime = nil
		}
		if elem.CustomerOwnedIpv4Pool != nil {
			cr.Spec.ForProvider.CustomerOwnedIPv4Pool = elem.CustomerOwnedIpv4Pool
		} else {
			cr.Spec.ForProvider.CustomerOwnedIPv4Pool = nil
		}
		if elem.DNSName != nil {
			cr.Status.AtProvider.DNSName = elem.DNSName
		} else {
			cr.Status.AtProvider.DNSName = nil
		}
		if elem.IpAddressType != nil {
			cr.Spec.ForProvider.IPAddressType = elem.IpAddressType
		} else {
			cr.Spec.ForProvider.IPAddressType = nil
		}
		if elem.LoadBalancerArn != nil {
			cr.Status.AtProvider.LoadBalancerARN = elem.LoadBalancerArn
		} else {
			cr.Status.AtProvider.LoadBalancerARN = nil
		}
		if elem.LoadBalancerName != nil {
			cr.Status.AtProvider.LoadBalancerName = elem.LoadBalancerName
		} else {
			cr.Status.AtProvider.LoadBalancerName = nil
		}
		if elem.Scheme != nil {
			cr.Spec.ForProvider.Scheme = elem.Scheme
		} else {
			cr.Spec.ForProvider.Scheme = nil
		}
		if elem.SecurityGroups != nil {
			f9 := []*string{}
			for _, f9iter := range elem.SecurityGroups {
				var f9elem string
				f9elem = *f9iter
				f9 = append(f9, &f9elem)
			}
			cr.Spec.ForProvider.SecurityGroups = f9
		} else {
			cr.Spec.ForProvider.SecurityGroups = nil
		}
		if elem.State != nil {
			f10 := &svcapitypes.LoadBalancerState{}
			if elem.State.Code != nil {
				f10.Code = elem.State.Code
			}
			if elem.State.Reason != nil {
				f10.Reason = elem.State.Reason
			}
			cr.Status.AtProvider.State = f10
		} else {
			cr.Status.AtProvider.State = nil
		}
		if elem.Type != nil {
			cr.Status.AtProvider.Type = elem.Type
		} else {
			cr.Status.AtProvider.Type = nil
		}
		if elem.VpcId != nil {
			cr.Status.AtProvider.VPCID = elem.VpcId
		} else {
			cr.Status.AtProvider.VPCID = nil
		}
		found = true
		break
	}
	if !found {
		_ = found
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	return e.update(ctx, mg)

}

func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.LoadBalancer)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteLoadBalancerInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return nil
	}
	resp, err := e.client.DeleteLoadBalancerWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.ELBV2API, opts []option) *external {
	e := &external{
		kube:           kube,
		client:         client,
		preObserve:     nopPreObserve,
		postObserve:    nopPostObserve,
		lateInitialize: nopLateInitialize,
		isUpToDate:     alwaysUpToDate,
		filterList:     nopFilterList,
		preCreate:      nopPreCreate,
		postCreate:     nopPostCreate,
		preDelete:      nopPreDelete,
		postDelete:     nopPostDelete,
		update:         nopUpdate,
	}
	for _, f := range opts {
		f(e)
	}
	return e
}

type external struct {
	kube           client.Client
	client         svcsdkapi.ELBV2API
	preObserve     func(context.Context, *svcapitypes.LoadBalancer, *svcsdk.DescribeLoadBalancersInput) error
	postObserve    func(context.Context, *svcapitypes.LoadBalancer, *svcsdk.DescribeLoadBalancersOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	filterList     func(*svcapitypes.LoadBalancer, *svcsdk.DescribeLoadBalancersOutput) *svcsdk.DescribeLoadBalancersOutput
	lateInitialize func(*svcapitypes.LoadBalancerParameters, *svcsdk.DescribeLoadBalancersOutput) error
	isUpToDate     func(*svcapitypes.LoadBalancer, *svcsdk.DescribeLoadBalancersOutput) (bool, error)
	preCreate      func(context.Context, *svcapitypes.LoadBalancer, *svcsdk.CreateLoadBalancerInput) error
	postCreate     func(context.Context, *svcapitypes.LoadBalancer, *svcsdk.CreateLoadBalancerOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete      func(context.Context, *svcapitypes.LoadBalancer, *svcsdk.DeleteLoadBalancerInput) (bool, error)
	postDelete     func(context.Context, *svcapitypes.LoadBalancer, *svcsdk.DeleteLoadBalancerOutput, error) error
	update         func(context.Context, cpresource.Managed) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.LoadBalancer, *svcsdk.DescribeLoadBalancersInput) error {
	return nil
}
func nopPostObserve(_ context.Context, _ *svcapitypes.LoadBalancer, _ *svcsdk.DescribeLoadBalancersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopFilterList(_ *svcapitypes.LoadBalancer, list *svcsdk.DescribeLoadBalancersOutput) *svcsdk.DescribeLoadBalancersOutput {
	return list
}

func nopLateInitialize(*svcapitypes.LoadBalancerParameters, *svcsdk.DescribeLoadBalancersOutput) error {
	return nil
}
func alwaysUpToDate(*svcapitypes.LoadBalancer, *svcsdk.DescribeLoadBalancersOutput) (bool, error) {
	return true, nil
}

func nopPreCreate(context.Context, *svcapitypes.LoadBalancer, *svcsdk.CreateLoadBalancerInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.LoadBalancer, _ *svcsdk.CreateLoadBalancerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopPreDelete(context.Context, *svcapitypes.LoadBalancer, *svcsdk.DeleteLoadBalancerInput) (bool, error) {
	return false, nil
}
func nopPostDelete(_ context.Context, _ *svcapitypes.LoadBalancer, _ *svcsdk.DeleteLoadBalancerOutput, err error) error {
	return err
}
func nopUpdate(context.Context, cpresource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}
