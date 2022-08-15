package exporter

import (
	"fmt"
	"testing"
)

func TestGetByCache(t *testing.T) {
	data, err := GetInstancesByCache()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(data))

	data, err = GetInstancesByCache()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(len(data))
}
