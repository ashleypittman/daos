#!/usr/bin/python
'''
  (C) Copyright 2019 Intel Corporation.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

  GOVERNMENT LICENSE RIGHTS-OPEN SOURCE SOFTWARE
  The Government's rights to use, modify, reproduce, release, perform, display,
  or disclose this software are subject to the terms of the Apache License as
  provided in Contract No. B609815.
  Any reproduction of computer software, computer software documentation, or
  portions thereof marked with this legend must also reproduce the markings.
'''

import os
#import sys
import io
import yaml

# a lookup table of predefined faults
FAULTS = {
    # this causes an object update to time out 17 times after
    # which it will start succeeding
    'obj_update_timeout': {
        'id': '50',
        'probability': '100',
        'interval': '1',
        'max_faults': '17'},
    # this doesn't exist in the code, delete me
    'bogus_2nd_fault': {
        'id': '51',
        'probability': '100',
        'interval': '1',
        'max_faults': '1'}
}

class FaultInjectionFailed(Exception):
    """Raise if FI failed"""

def WriteFaultFile(path, fault_list=None, on_the_fly_fault=None):
    """ Write out a fault injection config file.

        path             --where to write the file
        fault_list       --a list of strings identifying which predefined
                            faults to write out
        on_the_fly_fault --a fault dictionary that isn't predefined

        Returns the name of the file.
    """

    if fault_list is None and on_the_fly_fault is None:
        raise FaultInjectionFailed("bad parameters")

    if not os.path.exists(path):
        os.makedirs(path)
    fi_config = path + '/fi.yaml'

    with io.open(fi_config, 'w', encoding='utf8') as outfile:
        yaml.dump({'seed': '123'}, outfile, default_flow_style=False,
                  allow_unicode=True)

        fault_config = []
        if fault_list is not None:
            for fault in fault_list:
                fault_config.append(FAULTS[fault])
        if on_the_fly_fault is not None:
            fault_config.append(on_the_fly_fault)
        yaml.dump({'fault_config':fault_config}, outfile,
                  default_flow_style=False, allow_unicode=True)

    return fi_config

# just for testing above function
def main():
    on_the_fly = {'id': '52',
        'probability': '100',
        'interval': '1',
        'max_faults': '50'
    }
    print(WriteFaultFile('.', ['obj_update_timeout'], on_the_fly))

if __name__ == "__main__":
    main()
