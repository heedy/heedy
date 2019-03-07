package plugin

import (
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	ph := NewProcessHandler()

	//ph.StartCron("plugin1/proc1", "@every 5s", []string{"python", "-c", "print(\"running\")"})
	ph.StartProcess("plugin1/proc1", []string{"python", "-c", "import time;print(\"running\");time.sleep(10)"})
	d, _ := time.ParseDuration("5s")
	time.Sleep(d)
	ph.Stop(d)

}
