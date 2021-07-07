/**
 * (C) Copyright 2016-2021 Intel Corporation.
 *
 * SPDX-License-Identifier: BSD-2-Clause-Patent
 */

#include "dfuse_common.h"
#include "dfuse.h"

void
dfuse_cb_rename(fuse_req_t req, struct dfuse_inode_entry *parent,
		const char *name, struct dfuse_inode_entry *newparent,
		const char *newname, unsigned int flags)
{
	struct dfuse_projection_info *fs_handle;
	daos_obj_id_t	oid = {};
	int		rc;
	ino_t old_inode;
	struct dfuse_inode_entry *old_ie;
	bool inval_name = false;
	d_list_t			*rlink;

	fs_handle = fuse_req_userdata(req);

	if (flags != 0)
		D_GOTO(out, rc = ENOTSUP);

	if (!newparent)
		newparent = parent;

	rc = dfs_move(parent->ie_dfs->dfs_ns, parent->ie_obj, (char *)name,
		      newparent->ie_obj, (char *)newname, &oid);
	if (rc)
		D_GOTO(out, rc);

	DFUSE_TRA_INFO(parent, "Renamed %s to %s in %p",
		       name, newname, newparent);

	DFUSE_REPLY_ZERO(parent, req);

	/* If there is any caching then instruct the kernel to forget about whatever it thinks is
	 * at the new filename */
	if (parent->ie_dfs->dfc_dentry_timeout ||
		parent->ie_dfs->dfc_dentry_dir_timeout ||
		parent->ie_dfs->dfc_ndentry_timeout) {
		/* df_ll_rename() has a reference on newparent so it's safe to use it here */
		rc = fuse_lowlevel_notify_inval_entry(fs_handle->dpi_info->di_session,
						      newparent->ie_stat.st_ino,
						      newname, strnlen(newname, NAME_MAX));
		if (rc)
			DFUSE_TRA_ERROR(parent,
					"inval_entry on new name failed %d", rc);

		inval_name = true;
	}

	/* If there was no old file clobbered then do nothing */
	if (oid.hi == 0 && oid.lo == 0)
		return;

	dfuse_compute_inode(parent->ie_dfs, &oid, &old_inode);
	rlink = d_hash_rec_find(&fs_handle->dpi_iet, &old_inode, sizeof(old_inode));
	if (!rlink)
		return;

	old_ie = container_of(rlink, struct dfuse_inode_entry, ie_htl);

	/* Don't make the same inval call twice */
	if (inval_name && (newparent->ie_stat.st_ino == old_ie->ie_parent) &&
		(strncmp(newname, old_ie->ie_name, NAME_MAX) == 0))
		goto decref;

	rc = fuse_lowlevel_notify_inval_entry(fs_handle->dpi_info->di_session,	old_ie->ie_parent,
					      old_ie->ie_name, strnlen(old_ie->ie_name, NAME_MAX));
	if (rc)
		DFUSE_TRA_ERROR(parent,
				"inval_entry on old inode failed %d", rc);
decref:
	/* Drop the ref again */
	d_hash_rec_decref(&fs_handle->dpi_iet, rlink);

	return;

out:
	DFUSE_REPLY_ERR_RAW(parent, req, rc);
}
