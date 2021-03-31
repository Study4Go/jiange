package qconf

import (
	"testing"
)

func BenchmarkGetBatchKeys(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := GetBatchKeys("/arch_group/zkapi/arch.test.http/providers", "")
		if err != nil {
			b.Fatal("BenchmarkGetBatchKeys failed")
		}
	}
}
