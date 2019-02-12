//
// (C) Copyright 2018-2019 Intel Corporation.
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

#ifndef NVMECONTROL_H
#define NVMECONTROL_H

#include <stdbool.h>

#define NVMECONTROL_GBYTE_BYTES 1000000000

/**
 * \brief NVMe controller details
 */
struct ctrlr_t {
	int		id;
	char		model[1024];
	char		serial[1024];
	char		tr_addr[SPDK_NVMF_TRADDR_MAX_LEN + 1];
	char		fw_rev[1024];
	struct ctrlr_t	*next;
};

/**
 * \brief NVMe namespace details
 */
struct ns_t {
	int		id;
	int		size;
	char		ctrlr_pci_addr[1024];
	struct ns_t	*next;
};

/**
 * \brief Return containing return code, controllers, namespaces and error
 * message
 */
struct ret_t {
	int		rc;
	struct ctrlr_t	*ctrlrs;
	struct ns_t	*nss;
	char		err[1024];
};

/**
 * Discover NVMe controllers and namespaces.
 *
 * \return a pointer to a return struct (ret_t).
 */
struct ret_t *nvme_discover(void);

/**
 * Update NVMe controller firmware.
 *
 * \param ctrlr_pci_addr PCI address of NVMe controller.
 * \param path Local filepath where firmware image is stored.
 * \param slot Identifier of software slot/register to upload to.
 *
 * \return a pointer to a return struct (ret_t).
 */
struct ret_t *nvme_fwupdate(char *ctrlr_pci_addr, char *path, unsigned int slot);

/**
 * Cleanup structs held in memory.
 */
void nvme_cleanup(void);

#endif
