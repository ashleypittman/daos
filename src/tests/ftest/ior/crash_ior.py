#!/usr/bin/python
"""
  (C) Copyright 2020-2021 Intel Corporation.

  SPDX-License-Identifier: BSD-2-Clause-Patent
"""

import time

from ior_test_base import IorTestBase
from dmg_utils import check_system_query_status


class CrashIor(IorTestBase):
    # pylint: disable=too-many-ancestors
    """Test class Description: DAOS server does not need to be restarted
                               when the application crashes.
    :avocado: recursive
    """

    def setUp(self):
        """Set up test before executing."""
        super().setUp()
        self.dmg = self.get_dmg_command()

    def test_crashior(self):
        """Jira ID: DAOS-4332.

        Test Description:
            DAOS server does not need to be restarted when the application
            crashes.

        Use Cases:
            Run IOR over dfuse.
            Cancel IOR in the middle of io.
            Check daos server does not need to be restarted when the
            application crashes.

        :avocado: tags=all,daosio,hw,medium,ib2,full_regression,crashior
        """
        # run ior and crash it during write process
        self.run_ior_with_pool()
        # check if ior write has started
        self.check_subprocess_status()
        # allow 50 secs of write to happen
        time.sleep(50)
        # kill ior process in the middle of IO
        self.stop_ior()

        # obtain server rank info using 'dmg system query -v'
        scan_info = self.dmg.system_query(verbose=True)
        # check for any crashed servers after killing ior in the middle
        if not check_system_query_status(scan_info):
            self.fail("One or more server crashed")

        # run ior again and crash it during read process
        self.run_ior_with_pool()
        # allow write to finish which is set at stonewalling limit of 100 sec
        # hence allowing extra 5 secs for read to begin
        time.sleep(105)
        # check if ior read process started
        self.check_subprocess_status("read")
        # kill ior process in middle of read process
        self.stop_ior()

        # obtain server rank info using 'dmg system query -v'
        scan_info = self.dmg.system_query(verbose=True)
        # check for any crashed servers after killing ior in the middle
        if not check_system_query_status(scan_info):
            self.fail("One or more server crashed")

        # run ior again if everything goes well till now and allow it to
        # complete without killing in the middle this time to check
        # if io goes as expected after crashing it previously
        self.run_ior_with_pool()
        self.job_manager.wait()
