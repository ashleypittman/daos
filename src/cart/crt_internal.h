/*
 * (C) Copyright 2016-2024 Intel Corporation.
 * (C) Copyright 2025 Hewlett Packard Enterprise Development LP
 *
 * SPDX-License-Identifier: BSD-2-Clause-Patent
 */
/**
 * This file is part of CaRT. It it the common header file which be included by
 * all other .c files of CaRT.
 */

#ifndef __CRT_INTERNAL_H__
#define __CRT_INTERNAL_H__

#include "crt_debug.h"

#include <gurt/common.h>
#include <gurt/fault_inject.h>
#include <cart/api.h>

#include "crt_hg.h"
#include "crt_internal_types.h"
#include "crt_internal_fns.h"
#include "crt_rpc.h"
#include "crt_group.h"
#include "crt_tree.h"
#include "crt_self_test.h"
#include "crt_swim.h"

/* A wrapper around D_TRACE_DEBUG that ensures the ptr option is a RPC */
#define RPC_TRACE(mask, rpc, fmt, ...)                                                             \
	do {                                                                                       \
		char *_module;                                                                     \
		char *_opc;                                                                        \
                                                                                                   \
		if (!D_LOG_ENABLED(DB_TRACE))                                                      \
			break;                                                                     \
                                                                                                   \
		crt_opc_decode((rpc)->crp_pub.cr_opc, &_module, &_opc);                            \
		if ((rpc)->crp_coll) {                                                             \
			D_TRACE_DEBUG(mask, (rpc), "[opc=%#x (%s:%s) rpcid=%#lx CORPC] " fmt,      \
				      (rpc)->crp_pub.cr_opc, _module, _opc,                        \
				      (rpc)->crp_req_hdr.cch_rpcid, ##__VA_ARGS__);                \
		} else {                                                                           \
			D_TRACE_DEBUG(mask, (rpc),                                                 \
				      "[opc=%#x (%s:%s) rpcid=%#lx rank:tag=%d:%d] " fmt,          \
				      (rpc)->crp_pub.cr_opc, _module, _opc,                        \
				      (rpc)->crp_req_hdr.cch_rpcid, (rpc)->crp_pub.cr_ep.ep_rank,  \
				      (rpc)->crp_pub.cr_ep.ep_tag, ##__VA_ARGS__);                 \
		}                                                                                  \
	} while (0)

/* Log an error with an RPC descriptor */
#define RPC_ERROR(rpc, fmt, ...)                                                                   \
	do {                                                                                       \
		char *_module;                                                                     \
		char *_opc;                                                                        \
                                                                                                   \
		crt_opc_decode((rpc)->crp_pub.cr_opc, &_module, &_opc);                            \
		if ((rpc)->crp_coll) {                                                             \
			D_TRACE_ERROR((rpc), "[opc=%#x (%s:%s) rpcid=%#lx CORPC] " fmt,            \
				      (rpc)->crp_pub.cr_opc, _module, _opc,                        \
				      (rpc)->crp_req_hdr.cch_rpcid, ##__VA_ARGS__);                \
		} else {                                                                           \
			D_TRACE_ERROR((rpc), "[opc=%#x (%s:%s) rpcid=%#lx rank:tag=%d:%d] " fmt,   \
				      (rpc)->crp_pub.cr_opc, _module, _opc,                        \
				      (rpc)->crp_req_hdr.cch_rpcid, (rpc)->crp_pub.cr_ep.ep_rank,  \
				      (rpc)->crp_pub.cr_ep.ep_tag, ##__VA_ARGS__);                 \
		}                                                                                  \
	} while (0)

/* Log a warning with an RPC descriptor */
#define RPC_WARN(rpc, fmt, ...)                                                                    \
	do {                                                                                       \
		char *_module;                                                                     \
		char *_opc;                                                                        \
                                                                                                   \
		crt_opc_decode((rpc)->crp_pub.cr_opc, &_module, &_opc);                            \
		if ((rpc)->crp_coll) {                                                             \
			D_TRACE_WARN((rpc), "[opc=%#x (%s:%s) rpcid=%#lx CORPC] " fmt,             \
				     (rpc)->crp_pub.cr_opc, _module, _opc,                         \
				     (rpc)->crp_req_hdr.cch_rpcid, ##__VA_ARGS__);                 \
		} else {                                                                           \
			D_TRACE_WARN((rpc), "[opc=%#x (%s:%s) rpcid=%#lx rank:tag=%d:%d] " fmt,    \
				     (rpc)->crp_pub.cr_opc, _module, _opc,                         \
				     (rpc)->crp_req_hdr.cch_rpcid, (rpc)->crp_pub.cr_ep.ep_rank,   \
				     (rpc)->crp_pub.cr_ep.ep_tag, ##__VA_ARGS__);                  \
		}                                                                                  \
	} while (0)

/* Log an info message with an RPC descriptor */
#define RPC_INFO(rpc, fmt, ...)                                                                    \
	do {                                                                                       \
		char *_module;                                                                     \
		char *_opc;                                                                        \
                                                                                                   \
		crt_opc_decode((rpc)->crp_pub.cr_opc, &_module, &_opc);                            \
		if ((rpc)->crp_coll) {                                                             \
			D_TRACE_INFO((rpc), "[opc=%#x (%s:%s) rpcid=%#lx CORPC] " fmt,             \
				     (rpc)->crp_pub.cr_opc, _module, _opc,                         \
				     (rpc)->crp_req_hdr.cch_rpcid, ##__VA_ARGS__);                 \
		} else {                                                                           \
			D_TRACE_INFO((rpc), "[opc=%#x (%s:%s) rpcid=%#lx rank:tag=%d:%d] " fmt,    \
				     (rpc)->crp_pub.cr_opc, _module, _opc,                         \
				     (rpc)->crp_req_hdr.cch_rpcid, (rpc)->crp_pub.cr_ep.ep_rank,   \
				     (rpc)->crp_pub.cr_ep.ep_tag, ##__VA_ARGS__);                  \
		}                                                                                  \
	} while (0)
/**
 * If \a cond is false, this is equivalent to an RPC_ERROR (i.e., \a mask is
 * ignored). If \a cond is true, this is equivalent to an RPC_TRACE.
 */
#define RPC_CERROR(cond, mask, rpc, fmt, ...)				\
	do {								\
		if (cond)						\
			RPC_TRACE(mask, rpc, fmt, ## __VA_ARGS__);	\
		else							\
			RPC_ERROR(rpc, fmt, ## __VA_ARGS__);		\
	} while (0)

#define RPC_CWARN(cond, mask, rpc, fmt, ...)                                                       \
	do {                                                                                       \
		if (cond)                                                                          \
			RPC_TRACE(mask, rpc, fmt, ##__VA_ARGS__);                                  \
		else                                                                               \
			RPC_WARN(rpc, fmt, ##__VA_ARGS__);                                         \
	} while (0)

#ifdef CRT_DEBUG_TRACE
#	define CRT_ENTRY()					\
		D_DEBUG(DB_TRACE, ">>>> Entered %s: %d\n", __func__, __LINE__)

#	define CRT_EXIT()					\
		D_DEBUG(DB_TRACE, "<<<< Exit %s: %d\n", __func__, __LINE__)
#else
#	define CRT_ENTRY()	/* */
#	define CRT_EXIT()	/* */

#endif

/* crt uri lookup cache info */
struct crt_uri_cache {
	struct crt_grp_cache *grp_cache;
	uint32_t              max_count;
	uint32_t              idx;
};

void
crt_hdlr_ctl_get_uri_cache(crt_rpc_t *rpc_req);
void
crt_hdlr_ctl_ls(crt_rpc_t *rpc_req);
void
crt_hdlr_ctl_get_hostname(crt_rpc_t *rpc_req);
void
crt_hdlr_ctl_get_pid(crt_rpc_t *rpc_req);

void
crt_iv_init(crt_init_options_t *ops);
#endif /* __CRT_INTERNAL_H__ */
