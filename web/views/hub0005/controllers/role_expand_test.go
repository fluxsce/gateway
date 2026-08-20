package controllers

import (
	"sort"
	"testing"

	resourcemodels "gateway/web/views/hub0006/models"
)

func TestExpandSelectedResources_doesNotRestoreUncheckedChild(t *testing.T) {
	resources := []*resourcemodels.Resource{
		{ResourceId: "mod", ResourceCode: "hub0002", ParentResourceId: ""},
		{ResourceId: "search", ResourceCode: "hub0002:search", ParentResourceId: "mod"},
		{ResourceId: "add", ResourceCode: "hub0002:add", ParentResourceId: "mod"},
	}
	// 前端取消查询后只提交剩余按钮，模块由后端补父，不得把 search 加回去
	got := expandSelectedResources(map[string]bool{"add": true}, resources)
	sort.Strings(got)
	want := []string{"add", "mod"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestExpandSelectedResources_checkingModuleFillsButtons(t *testing.T) {
	resources := []*resourcemodels.Resource{
		{ResourceId: "mod", ResourceCode: "hub0002", ParentResourceId: ""},
		{ResourceId: "search", ResourceCode: "hub0002:search", ParentResourceId: "mod"},
		{ResourceId: "add", ResourceCode: "hub0002:add", ParentResourceId: "mod"},
	}
	got := expandSelectedResources(map[string]bool{"mod": true}, resources)
	sort.Strings(got)
	want := []string{"add", "mod", "search"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
