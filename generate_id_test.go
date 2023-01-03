package xl

import "testing"

func TestGenerateId(t *testing.T) {
	id := GenerateUUID()
	t.Log(id)
}
