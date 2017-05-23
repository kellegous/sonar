package web

import (
	"testing"
)

func TestPerc(t *testing.T) {
	perc(0.1, []int{1})
	perc(0.5, []int{1})
	perc(0.9, []int{1})
}
