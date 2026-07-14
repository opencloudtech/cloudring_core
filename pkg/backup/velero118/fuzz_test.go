// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package velero118

import "testing"

func FuzzDecodeDataDownload(f *testing.F) {
	f.Add([]byte(`{"apiVersion":"velero.io/v2alpha1","kind":"DataDownload","metadata":{"name":"one","namespace":"velero","uid":"uid","resourceVersion":"1"},"spec":{},"status":{}}`))
	f.Add([]byte(`{"apiVersion":"velero.io/v2alpha1","kind":"DataDownload","metadata":{"name":"one","name":"two"}}`))
	f.Fuzz(func(t *testing.T, value []byte) {
		_, _ = DecodeDataDownload(value)
	})
}

func FuzzDecodeListPage(f *testing.F) {
	f.Add([]byte(`{"apiVersion":"v1","kind":"ConfigMapList","metadata":{"resourceVersion":"1","continue":""},"items":[]}`))
	f.Add([]byte(`{"kind":"List","items":[]}`))
	f.Fuzz(func(t *testing.T, value []byte) {
		_, _ = DecodeListPage(value, "v1", "ConfigMapList")
	})
}
