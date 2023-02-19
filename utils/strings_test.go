package utils

import "testing"

func TestToSnakeCase(t *testing.T) {
	v := "HelloWorld"
	if ToSnakeCase(v) != "hello_world" {
		t.Fatal("HelloWorld failed")
	}
	v = "ID"
	if ToSnakeCase(v) != "id" {
		t.Fatal("ID failed")
	}
}
