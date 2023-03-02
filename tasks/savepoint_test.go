package tasks

import (
	"context"
	"testing"
)

func TestSavepoint(t *testing.T) {
	fs := FileSavepoint("/tmp/savepoint")
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for i := 1000000; i > 0; i /= 10 {
		if err := fs.SetOffset(ctx, i); err != nil {
			t.Fatal(err)
		}
		var offset int
		if err := fs.Offset(ctx, &offset); err != nil {
			t.Fatal(err)
		}
		if offset != i {
			t.Fatal("failed")
		}
	}
}
