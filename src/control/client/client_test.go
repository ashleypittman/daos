//
// (C) Copyright 2018 Intel Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// GOVERNMENT LICENSE RIGHTS-OPEN SOURCE SOFTWARE
// The Government's rights to use, modify, reproduce, release, perform, display,
// or disclose this software are subject to the terms of the Apache License as
// provided in Contract No. 8F-30005.
// Any reproduction of computer software, computer software documentation, or
// portions thereof marked with this legend must also reproduce the markings.
//

package client

import (
	"fmt"

	pb "github.com/daos-stack/daos/src/control/common/proto/mgmt"
	"google.golang.org/grpc/connectivity"
)

// implement mock/stub behaviour for Control
type mockControl struct {
	address   string
	connState connectivity.State
	features  []*pb.Feature
	ctrlrs    NvmeControllers
	modules   ScmModules
}

func (m *mockControl) connect(addr string) error {
	m.address = addr
	return nil
}
func (m *mockControl) close() error { return nil }
func (m *mockControl) connected() (connectivity.State, bool) {
	return m.connState, checkState(m.connState)
}
func (m *mockControl) getAddress() string { return m.address }
func (m *mockControl) listAllFeatures() (FeatureMap, error) {
	fm := make(FeatureMap)
	for _, f := range m.features {
		fm[f.Fname.Name] = fmt.Sprintf(
			"category %s, %s", f.Category.Category, f.Description)
	}
	return fm, nil
}
func (m *mockControl) listNvmeCtrlrs() (NvmeControllers, error) {
	return m.ctrlrs, nil
}
func (m *mockControl) listScmModules() (ScmModules, error) {
	return m.modules, nil
}

func newMockControl(
	address string, state connectivity.State, features []*pb.Feature,
	ctrlrs NvmeControllers, modules ScmModules) Control {

	return &mockControl{address, state, features, ctrlrs, modules}
}
