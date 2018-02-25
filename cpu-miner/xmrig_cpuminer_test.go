package cpuminer

import (
	"testing"
)

func TestXMRigCPUMiner(t *testing.T) {
	testCPUMiner(t, 4, NewXMRigCPUMiner)
}
