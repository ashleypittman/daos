//
// (C) Copyright 2021-2024 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

package main

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	"github.com/daos-stack/daos/src/control/common"
	"github.com/daos-stack/daos/src/control/common/test"
	"github.com/daos-stack/daos/src/control/lib/hardware"
	"github.com/daos-stack/daos/src/control/logging"
)

var fiCmpOpt = cmpopts.IgnoreUnexported(FabricInterface{})

func testFabricProviderSet(prov ...string) *hardware.FabricProviderSet {
	providers := []*hardware.FabricProvider{}
	for _, p := range prov {
		providers = append(providers, &hardware.FabricProvider{
			Name: p,
		})
	}
	return hardware.NewFabricProviderSet(providers...)
}

func TestAgent_NUMAFabricMap_MaxNUMANode(t *testing.T) {
	for name, tc := range map[string]struct {
		nfm       NUMAFabricMap
		expResult int
	}{
		"nil": {
			expResult: -1,
		},
		"empty": {
			nfm:       NUMAFabricMap{},
			expResult: -1,
		},
		"single node 0": {
			nfm: NUMAFabricMap{
				0: []*FabricInterface{},
			},
			expResult: 0,
		},
		"consecutive": {
			nfm: NUMAFabricMap{
				0: []*FabricInterface{},
				1: []*FabricInterface{},
				2: []*FabricInterface{},
			},
			expResult: 2,
		},
		"non-consecutive": {
			nfm: NUMAFabricMap{
				2: []*FabricInterface{},
				7: []*FabricInterface{},
			},
			expResult: 7,
		},
	} {
		t.Run(name, func(t *testing.T) {
			test.AssertEqual(t, tc.expResult, tc.nfm.MaxNUMANode(), "")
		})
	}
}

func TestAgent_NewNUMAFabric(t *testing.T) {
	log, buf := logging.NewTestLogger(t.Name())
	defer test.ShowBufferOnFailure(t, buf)

	result := newNUMAFabric(log)

	if result == nil {
		t.Fatal("result was nil")
	}

	if result.numaMap == nil {
		t.Fatal("map was nil")
	}
}

func TestAgent_NUMAFabric_NumNUMANodes(t *testing.T) {
	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		expResult int
	}{
		"nil": {},
		"empty struct": {
			nf: &NUMAFabric{},
		},
		"multiple nodes": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0:  {},
					1:  {},
					3:  {},
					10: {},
				},
			},
			expResult: 4,
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)
			if tc.nf != nil {
				tc.nf.log = log
			}

			test.AssertEqual(t, tc.expResult, tc.nf.NumNUMANodes(), "")
		})
	}
}

func TestAgent_NUMAFabric_NumDevices(t *testing.T) {
	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		node      int
		expResult int
	}{
		"nil": {},
		"empty": {
			nf:   &NUMAFabric{},
			node: 5,
		},
		"multiple devices on node": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					3: {&FabricInterface{}, &FabricInterface{}},
				},
			},
			node:      3,
			expResult: 2,
		},
		"multiple nodes": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0:  {&FabricInterface{}, &FabricInterface{}, &FabricInterface{}},
					1:  {&FabricInterface{}},
					3:  {&FabricInterface{}},
					10: {&FabricInterface{}},
				},
			},
			node:      0,
			expResult: 3,
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)
			if tc.nf != nil {
				tc.nf.log = log
			}

			test.AssertEqual(t, tc.expResult, tc.nf.NumDevices(tc.node), "")
		})
	}
}

func TestAgent_NUMAFabric_Add(t *testing.T) {
	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		input     *FabricInterface
		node      int
		expErr    error
		expResult NUMAFabricMap
	}{
		"nil": {
			expErr: errors.New("nil NUMAFabric"),
		},
		"empty": {
			nf:    newNUMAFabric(nil),
			input: &FabricInterface{Name: "test1"},
			node:  2,
			expResult: NUMAFabricMap{
				2: {{Name: "test1"}},
			},
		},
		"non-empty": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					2: {{Name: "t1"}},
					3: {{Name: "t2"}, {Name: "t3"}},
				},
			},
			input: &FabricInterface{Name: "test1"},
			node:  2,
			expResult: NUMAFabricMap{
				2: {{Name: "t1"}, {Name: "test1"}},
				3: {{Name: "t2"}, {Name: "t3"}},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)
			if tc.nf != nil {
				tc.nf.log = log
			}

			err := tc.nf.Add(tc.node, tc.input)

			test.CmpErr(t, tc.expErr, err)

			if tc.nf == nil {
				return
			}
			if diff := cmp.Diff(tc.expResult, tc.nf.numaMap, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}

func TestAgent_NUMAFabric_GetDevice(t *testing.T) {
	for name, tc := range map[string]struct {
		nf         *NUMAFabric
		params     *FabricIfaceParams
		include    []string
		exclude    []string
		expErr     error
		expResults []*FabricInterface
	}{
		"nil": {
			expErr: errors.New("nil NUMAFabric"),
		},
		"nil params": {
			nf:     newNUMAFabric(nil),
			expErr: errors.New("nil FabricIfaceParams"),
		},
		"empty": {
			nf: newNUMAFabric(nil),
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Loopback,
			},
			expErr: errors.New("no suitable fabric interface"),
		},
		"no provider": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t3"),
							Name:          "t3",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				DevClass: hardware.Ether,
			},
			expErr: errors.New("provider is required"),
		},
		"type not found": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t3"),
							Name:          "t3",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Infiniband,
			},
			expErr: errors.New("no suitable fabric interface"),
		},
		"provider not found": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Infiniband,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Infiniband,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t3"),
							Name:          "t3",
							DeviceClass:   hardware.Infiniband,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+verbs",
				DevClass: hardware.Infiniband,
			},
			expErr: errors.New("no suitable fabric interface"),
		},
		"choose first device": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t1"),
						Name:          "t1",
						DeviceClass:   hardware.Infiniband,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Infiniband,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t1",
					Domain:      "t1",
					NetDevClass: hardware.Infiniband,
				},
				{
					Name:        "t1",
					Domain:      "t1",
					NetDevClass: hardware.Infiniband,
				},
			},
		},
		"choose later device": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Infiniband,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Infiniband,
				NUMANode: 0,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Infiniband,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Infiniband,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Infiniband,
				},
			},
		},
		"nothing on NUMA node": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t1"),
						Name:          "t1",
						DeviceClass:   hardware.Infiniband,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
					1: {},
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Infiniband,
				NUMANode: 1,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t1",
					Domain:      "t1",
					NetDevClass: hardware.Infiniband,
				},
				{
					Name:        "t1",
					Domain:      "t1",
					NetDevClass: hardware.Infiniband,
				},
			},
		},
		"type not found on NUMA node": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t1"),
						Name:          "t1",
						DeviceClass:   hardware.Infiniband,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
					3: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t2"),
						Name:          "t2",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Ether,
				NUMANode: 1,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"manual FI matches any": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: FabricDevClassManual,
						},
					},
					1: {
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: FabricDevClassManual,
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Infiniband,
				NUMANode: 1,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: FabricDevClassManual,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: FabricDevClassManual,
				},
			},
		},
		"load balancing on NUMA node": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t3"),
							Name:          "t3",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
				currentNumaDevIdx: map[int]int{
					0: 1,
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Ether,
				NUMANode: 0,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t3",
					Domain:      "t3",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t1",
					Domain:      "t1",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"load balancing amongst NUMA nodes": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t1"),
						Name:          "t1",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
					1: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t2"),
						Name:          "t2",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
					2: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t3"),
						Name:          "t3",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Ether,
				NUMANode: 3,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t3",
					Domain:      "t3",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t1",
					Domain:      "t1",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"validating IPs fails": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t1"),
						Name:          "t1",
						DeviceClass:   hardware.Infiniband,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
				},
				getAddrInterface: func(_ string) (addrFI, error) {
					return nil, errors.New("mock getAddrInterface")
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+sockets",
				DevClass: hardware.Infiniband,
				NUMANode: 0,
			},
			expErr: FabricNotFoundErr(hardware.Infiniband),
		},
		"specific provider": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2_dom",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets", "ofi+verbs"),
						})[0],
					},
					1: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t3"),
						Name:          "t3_dom",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+verbs",
				DevClass: hardware.Ether,
				NUMANode: 0,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2_dom",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2_dom",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2_dom",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"specific provider from other numa": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t1"),
						Name:          "t1",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+sockets"),
					}),
					1: fabricInterfacesFromHardware(&hardware.FabricInterface{
						NetInterfaces: common.NewStringSet("t2"),
						Name:          "t2_dom",
						DeviceClass:   hardware.Ether,
						Providers:     testFabricProviderSet("ofi+verbs"),
					}),
				},
			},
			params: &FabricIfaceParams{
				Provider: "ofi+verbs",
				DevClass: hardware.Ether,
				NUMANode: 0,
			},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2_dom",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2_dom",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"include interface": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
					1: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				NUMANode: 0,
				Provider: "ofi+sockets",
				DevClass: hardware.Ether,
			},
			include: []string{"t2"},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"exclude interface": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
					1: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				NUMANode: 0,
				Provider: "ofi+sockets",
				DevClass: hardware.Ether,
			},
			exclude: []string{"t1"},
			expResults: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"exclude all interfaces": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t1"),
							Name:          "t1",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
						fabricInterfacesFromHardware(&hardware.FabricInterface{
							NetInterfaces: common.NewStringSet("t2"),
							Name:          "t2",
							DeviceClass:   hardware.Ether,
							Providers:     testFabricProviderSet("ofi+sockets"),
						})[0],
					},
				},
			},
			params: &FabricIfaceParams{
				NUMANode: 0,
				Provider: "ofi+sockets",
				DevClass: hardware.Ether,
			},
			exclude: []string{"t1", "t2"},
			expErr:  errors.New("no suitable fabric interface"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)
			if tc.nf != nil {
				tc.nf.log = log
				if tc.nf.getAddrInterface == nil {
					tc.nf.getAddrInterface = getMockNetInterfaceSuccess
				}

				mode := filterModeExclude
				devSet := common.NewStringSet(tc.exclude...)
				if len(tc.include) > 0 {
					mode = filterModeInclude
					devSet = common.NewStringSet(tc.include...)
				}
				tc.nf = tc.nf.WithDeviceFilter(newDeviceFilter(devSet, mode))
			}

			numDevices := 0
			if tc.params != nil {
				numDevices = tc.nf.NumDevices(tc.params.NUMANode)
			}

			if numDevices == 0 && tc.nf != nil {
				for numa := range tc.nf.numaMap {
					numDevices += tc.nf.NumDevices(numa)
				}
			}

			var results []*FabricInterface
			for i := 0; i < numDevices+1; i++ {
				result, err := tc.nf.GetDevice(tc.params)
				test.CmpErr(t, tc.expErr, err)
				if tc.expErr != nil {
					return
				}
				results = append(results, result)
			}

			if diff := cmp.Diff(tc.expResults, results, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}

func TestAgent_NUMAFabric_Find(t *testing.T) {
	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		name      string
		expResult []*FabricInterface
		expErr    error
	}{
		"nil": {
			name:   "eth0",
			expErr: errors.New("nil"),
		},
		"not found": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t2",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t3",
							NetDevClass: hardware.Ether,
						},
					},
				},
			},
			name:   "t4",
			expErr: errors.New("not found"),
		},
		"found": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t2",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t3",
							NetDevClass: hardware.Ether,
						},
					},
				},
			},
			name: "t2",
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					NetDevClass: hardware.Ether,
				},
			},
		},
		"multiple": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t2",
							NetDevClass: hardware.Infiniband,
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
						},
					},
				},
			},
			name: "t2",
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					NetDevClass: hardware.Infiniband,
				},
				{
					Name:        "t2",
					Domain:      "d2",
					NetDevClass: hardware.Infiniband,
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := tc.nf.Find(tc.name)

			test.CmpErr(t, tc.expErr, err)
			if diff := cmp.Diff(tc.expResult, result, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}

func TestAgent_NUMAFabric_FindDevice(t *testing.T) {
	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		params    *FabricIfaceParams
		expResult []*FabricInterface
		expErr    error
	}{
		"nil": {
			params: &FabricIfaceParams{
				Interface: "eth0",
			},
			expErr: errors.New("nil"),
		},
		"nil params": {
			nf:     newNUMAFabric(nil),
			expErr: errors.New("nil"),
		},
		"name not found": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t2",
							NetDevClass: hardware.Ether,
						},
						{
							Name:        "t3",
							NetDevClass: hardware.Ether,
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t4",
			},
			expErr: errors.New("not found"),
		},
		"no domain match": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Domain:    "d1",
				Provider:  "p1",
			},
			expErr: errors.New("doesn't have requested domain"),
		},
		"no provider match": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Domain:    "d2",
				Provider:  "p2",
			},
			expErr: errors.New("doesn't support provider"),
		},
		"success": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Domain:    "d2",
				Provider:  "p2",
			},
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "d2",
					NetDevClass: hardware.Infiniband,
					hw: &hardware.FabricInterface{
						Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
					},
				},
			},
		},
		"success with manual interfaces": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: FabricDevClassManual,
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: FabricDevClassManual,
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Domain:    "t2",
				Provider:  "p2",
			},
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: FabricDevClassManual,
				},
			},
		},
		"success with no domain": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Provider:  "p2",
			},
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "d2",
					NetDevClass: hardware.Infiniband,
					hw: &hardware.FabricInterface{
						Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
					},
				},
			},
		},
		"domain is name": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Domain:    "t2",
				Provider:  "p1",
			},
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "",
					NetDevClass: hardware.Infiniband,
					hw: &hardware.FabricInterface{
						Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
					},
				},
			},
		},
		"success with no provider": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
				Domain:    "d2",
			},
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "d2",
					NetDevClass: hardware.Infiniband,
					hw: &hardware.FabricInterface{
						Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
					},
				},
			},
		},
		"more than one match": {
			nf: &NUMAFabric{
				numaMap: map[int][]*FabricInterface{
					0: {
						{
							Name:        "t1",
							Domain:      "t1",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "t2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
							},
						},
						{
							Name:        "t2",
							Domain:      "d2",
							NetDevClass: hardware.Infiniband,
							hw: &hardware.FabricInterface{
								Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
							},
						},
					},
				},
			},
			params: &FabricIfaceParams{
				Interface: "t2",
			},
			expResult: []*FabricInterface{
				{
					Name:        "t2",
					Domain:      "t2",
					NetDevClass: hardware.Infiniband,
					hw: &hardware.FabricInterface{
						Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
					},
				},
				{
					Name:        "t2",
					Domain:      "d2",
					NetDevClass: hardware.Infiniband,
					hw: &hardware.FabricInterface{
						Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p2"}),
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := tc.nf.FindDevice(tc.params)

			test.CmpErr(t, tc.expErr, err)
			if diff := cmp.Diff(tc.expResult, result, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}

func TestAgent_NUMAFabric_RLockedMap(t *testing.T) {
	fullMap := NUMAFabricMap{
		0: {
			{
				Name:        "t1",
				Domain:      "t1",
				NetDevClass: hardware.Infiniband,
				hw: &hardware.FabricInterface{
					Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
				},
			},
			{
				Name:        "t2",
				Domain:      "t2",
				NetDevClass: hardware.Infiniband,
				hw: &hardware.FabricInterface{
					Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
				},
			},
			{
				Name:        "t2",
				Domain:      "d2",
				NetDevClass: hardware.Infiniband,
				hw: &hardware.FabricInterface{
					Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
				},
			},
		},
	}

	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		expResult NUMAFabricMap
		expFunc   bool
		expErr    error
	}{
		"nil": {
			expErr: errors.New("nil"),
		},
		"nil map": {
			nf:     &NUMAFabric{},
			expErr: errors.New("uninitialized"),
		},
		"success": {
			nf: &NUMAFabric{
				numaMap: fullMap,
			},
			expResult: fullMap,
			expFunc:   true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, release, err := tc.nf.RLockedMap()
			if release != nil {
				defer release()
			}

			test.CmpErr(t, tc.expErr, err)
			if diff := cmp.Diff(tc.expResult, result, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}

			test.AssertEqual(t, tc.expFunc, release != nil, "expected release function")
		})
	}
}

func TestAgent_NUMAFabric_LockedMap(t *testing.T) {
	fullMap := NUMAFabricMap{
		0: {
			{
				Name:        "t1",
				Domain:      "t1",
				NetDevClass: hardware.Infiniband,
				hw: &hardware.FabricInterface{
					Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
				},
			},
			{
				Name:        "t2",
				Domain:      "t2",
				NetDevClass: hardware.Infiniband,
				hw: &hardware.FabricInterface{
					Providers: hardware.NewFabricProviderSet(&hardware.FabricProvider{Name: "p1"}),
				},
			},
		},
	}

	for name, tc := range map[string]struct {
		nf        *NUMAFabric
		expResult NUMAFabricMap
		expFunc   bool
		expErr    error
	}{
		"nil": {
			expErr: errors.New("nil"),
		},
		"nil map": {
			nf:     &NUMAFabric{},
			expErr: errors.New("uninitialized"),
		},
		"success": {
			nf: &NUMAFabric{
				numaMap: fullMap,
			},
			expResult: fullMap,
			expFunc:   true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			result, release, err := tc.nf.LockedMap()
			if release != nil {
				defer release()
			}

			test.CmpErr(t, tc.expErr, err)
			if diff := cmp.Diff(tc.expResult, result, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}

			test.AssertEqual(t, tc.expFunc, release != nil, "expected release function")
		})
	}
}

func TestAgent_NUMAFabricFromScan(t *testing.T) {
	for name, tc := range map[string]struct {
		input     *hardware.FabricInterfaceSet
		expResult NUMAFabricMap
	}{
		"no devices in scan": {
			expResult: NUMAFabricMap{},
		},
		"include lo": {
			input: hardware.NewFabricInterfaceSet(
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+sockets"),
					Name:          "test0",
					NetInterfaces: common.NewStringSet("os_test0"),
					NUMANode:      1,
					DeviceClass:   hardware.Ether,
				},
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+sockets"),
					Name:          "lo",
					NetInterfaces: common.NewStringSet("lo"),
					NUMANode:      1,
					DeviceClass:   hardware.Loopback,
				},
			),
			expResult: NUMAFabricMap{
				1: {

					{
						Name:        "lo",
						Domain:      "lo",
						NetDevClass: hardware.Loopback,
					},
					{
						Name:        "os_test0",
						Domain:      "test0",
						NetDevClass: hardware.Ether,
					},
				},
			},
		},
		"multiple devices": {
			input: hardware.NewFabricInterfaceSet(
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+sockets"),
					Name:          "test0",
					NetInterfaces: common.NewStringSet("os_test0"),
					NUMANode:      1,
					DeviceClass:   hardware.Ether,
				},
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+verbs"),
					Name:          "test1",
					NetInterfaces: common.NewStringSet("os_test1"),
					NUMANode:      0,
					DeviceClass:   hardware.Infiniband,
				},
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+sockets"),
					Name:          "test2",
					NetInterfaces: common.NewStringSet("os_test2"),
					NUMANode:      0,
					DeviceClass:   hardware.Ether,
				},
			),
			expResult: NUMAFabricMap{
				0: {
					{
						Name:        "os_test1",
						Domain:      "test1",
						NetDevClass: hardware.Infiniband,
					},
					{
						Name:        "os_test2",
						Domain:      "test2",
						NetDevClass: hardware.Ether,
					},
				},
				1: {
					{
						Name:        "os_test0",
						Domain:      "test0",
						NetDevClass: hardware.Ether,
					},
				},
			},
		},
		"multiple providers per device": {
			input: hardware.NewFabricInterfaceSet(
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+sockets", "ofi+tcp"),
					Name:          "test0",
					NetInterfaces: common.NewStringSet("os_test0"),
					NUMANode:      1,
					DeviceClass:   hardware.Ether,
				},
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+verbs"),
					Name:          "verbs_test1",
					NetInterfaces: common.NewStringSet("os_test1"),
					NUMANode:      0,
					DeviceClass:   hardware.Infiniband,
				},
				&hardware.FabricInterface{
					Providers:     testFabricProviderSet("ofi+sockets", "ofi+tcp"),
					Name:          "test1",
					NetInterfaces: common.NewStringSet("os_test1"),
					NUMANode:      0,
					DeviceClass:   hardware.Infiniband,
				},
			),
			expResult: NUMAFabricMap{
				0: {

					{
						Name:        "os_test1",
						Domain:      "test1",
						NetDevClass: hardware.Infiniband,
					},
					{
						Name:        "os_test1",
						Domain:      "verbs_test1",
						NetDevClass: hardware.Infiniband,
					},
				},
				1: {
					{
						Name:        "os_test0",
						Domain:      "test0",
						NetDevClass: hardware.Ether,
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)

			result := NUMAFabricFromScan(test.Context(t), log, tc.input)

			if diff := cmp.Diff(tc.expResult, result.numaMap, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}

func TestAgent_NUMAFabricFromConfig(t *testing.T) {
	for name, tc := range map[string]struct {
		input     []*NUMAFabricConfig
		expResult NUMAFabricMap
	}{
		"empty input": {
			expResult: NUMAFabricMap{},
		},
		"no devices on NUMA node": {
			input: []*NUMAFabricConfig{
				{
					NUMANode:   1,
					Interfaces: []*FabricInterfaceConfig{},
				},
			},
			expResult: NUMAFabricMap{},
		},
		"single NUMA node": {
			input: []*NUMAFabricConfig{
				{
					NUMANode: 1,
					Interfaces: []*FabricInterfaceConfig{
						{
							Interface: "test0",
							Domain:    "test0_domain",
						},
					},
				},
			},
			expResult: NUMAFabricMap{
				1: {
					{
						Name:        "test0",
						Domain:      "test0_domain",
						NetDevClass: FabricDevClassManual,
					},
				},
			},
		},
		"multiple devices": {
			input: []*NUMAFabricConfig{
				{
					NUMANode: 0,
					Interfaces: []*FabricInterfaceConfig{
						{
							Interface: "test1",
						},
						{
							Interface: "test2",
							Domain:    "test2_domain",
						},
					},
				},
				{
					NUMANode: 1,
					Interfaces: []*FabricInterfaceConfig{
						{
							Interface: "test0",
							Domain:    "test0_domain",
						},
					},
				},
			},
			expResult: NUMAFabricMap{
				0: {
					{
						Name:        "test1",
						Domain:      "test1",
						NetDevClass: FabricDevClassManual,
					},
					{
						Name:        "test2",
						Domain:      "test2_domain",
						NetDevClass: FabricDevClassManual,
					},
				},
				1: {
					{
						Name:        "test0",
						Domain:      "test0_domain",
						NetDevClass: FabricDevClassManual,
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)

			result := NUMAFabricFromConfig(log, tc.input)

			if diff := cmp.Diff(tc.expResult, result.numaMap, fiCmpOpt); diff != "" {
				t.Fatalf("-want, +got:\n%s", diff)
			}
		})
	}
}

type mockNetInterface struct {
	addrs    []net.Addr
	addrsErr error
}

func (m *mockNetInterface) Addrs() ([]net.Addr, error) {
	return m.addrs, m.addrsErr
}

func getMockNetInterfaceSuccess(_ string) (addrFI, error) {
	return &mockNetInterface{
		addrs: []net.Addr{
			&net.IPNet{
				IP: net.IP("127.0.0.1"),
			},
		},
	}, nil
}

func TestAgent_NUMAFabric_validateDevice(t *testing.T) {
	getMockNetInterfaceFunc := func(addrs []net.Addr, err error) func(string) (addrFI, error) {
		return func(_ string) (addrFI, error) {
			return &mockNetInterface{
				addrs:    addrs,
				addrsErr: err,
			}, nil
		}
	}

	for name, tc := range map[string]struct {
		getAddrInterface func(name string) (addrFI, error)
		expErr           error
	}{
		"getAddrInterface fails": {
			getAddrInterface: func(name string) (addrFI, error) {
				return nil, errors.New("mock getAddrInterface")
			},
			expErr: errors.New("mock getAddrInterface"),
		},
		"interface Addrs() fails": {
			getAddrInterface: getMockNetInterfaceFunc(nil, errors.New("mock Addrs()")),
			expErr:           errors.New("mock Addrs()"),
		},
		"empty Addrs()": {
			getAddrInterface: getMockNetInterfaceFunc([]net.Addr{}, nil),
			expErr:           errors.New("no IP addresses"),
		},
		"no IP addrs": {
			getAddrInterface: getMockNetInterfaceFunc([]net.Addr{
				&net.TCPAddr{},
			}, nil),
			expErr: errors.New("no IP addresses"),
		},
		"IP addr is empty": {
			getAddrInterface: getMockNetInterfaceFunc([]net.Addr{
				&net.IPNet{},
			}, nil),
			expErr: errors.New("no IP addresses"),
		},
		"IP addr is unspecified": {
			getAddrInterface: getMockNetInterfaceFunc([]net.Addr{
				&net.IPNet{
					IP: net.IPv4zero,
				},
			}, nil),
			expErr: errors.New("no IP addresses"),
		},
		"success": {
			getAddrInterface: getMockNetInterfaceSuccess,
		},
	} {
		t.Run(name, func(t *testing.T) {
			log, buf := logging.NewTestLogger(t.Name())
			defer test.ShowBufferOnFailure(t, buf)

			nf := newNUMAFabric(log)
			nf.getAddrInterface = tc.getAddrInterface

			err := nf.validateDevice(&FabricInterface{
				Name: "not_real",
			})

			test.CmpErr(t, tc.expErr, err)
		})
	}
}
