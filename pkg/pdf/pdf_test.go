package pdf

import (
	"reflect"
	"testing"
)

func TestGetParitionMap(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	paritionMap := getParitionMap(len(list), 2)
	expectedParitionMap := [][]int{{0, 2}, {2, 4}, {4, 5}}
	if !reflect.DeepEqual(paritionMap, expectedParitionMap) {
		t.Errorf("Partition was incorrect, got: %v, want: %v.", paritionMap, expectedParitionMap)
	}
}
