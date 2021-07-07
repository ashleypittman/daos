/**
 * (C) Copyright 2016-2021 Intel Corporation.
 *
 * SPDX-License-Identifier: BSD-2-Clause-Patent
 */

#include "dfuse_common.h"
#include "dfuse.h"

void
dfuse_cb_unlink(fuse_req_t req, struct dfuse_inode_entry *parent,
		const char *name)
{
	struct dfuse_projection_info *fs_handle;
	int		rc;
	daos_obj_id_t	oid = {};
	ino_t old_inode;
	struct dfuse_inode_entry *old_ie;
	d_list_t			*rlink;

	fs_handle = fuse_req_userdata(req);

	rc = dfs_remove(parent->ie_dfs->dfs_ns, parent->ie_obj, name, false, &oid);
	if (rc != 0) {
		DFUSE_REPLY_ERR_RAW(parent, req, rc);
		return;
	}

	DFUSE_REPLY_ZERO(parent, req);

	dfuse_compute_inode(parent->ie_dfs, &oid, &old_inode);
	rlink = d_hash_rec_find(&fs_handle->dpi_iet, &old_inode, sizeof(old_inode));
	if (!rlink)
		return;

	old_ie = container_of(rlink, struct dfuse_inode_entry, ie_htl);

	rc = fuse_lowlevel_notify_inval_entry(fs_handle->dpi_info->di_session,	old_ie->ie_parent,
					old_ie->ie_name, strnlen(old_ie->ie_name, NAME_MAX));
	if (rc)
		DFUSE_TRA_ERROR(parent,
				"inval_entry on old inode failed %d", rc);
	/* Drop the ref again */
	d_hash_rec_decref(&fs_handle->dpi_iet, rlink);
}
