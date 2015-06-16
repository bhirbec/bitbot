package config

import (
	"testing"
)

func TestAll(t *testing.T) {
	config := NewConfig()
	testString := config.String("main", "teststring")
	config.Load("config_test.ini")

	if *testString != "mystring" {
		t.Errorf("testString should be `teststring` instead of `%s`", *testString)
	}
}

func TestMissing(t *testing.T) {
	config := NewConfig()
	config.String("main", "missing")
	err := config.Load("config_test.ini")

	if err == nil {
		t.Error("Missing entry should return an error (got nil instead)")
	}
}
