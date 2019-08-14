package apix

import (
	"fmt"
	"testing"
)

func TestNewServer(t *testing.T) {
	var args []interface{} = make([]interface{}, 0)
	test(args...)
}

func test(args ...interface{}) {
	fmt.Printf("%v", args == nil)
}
