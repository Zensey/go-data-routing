package go_data_routing

import (
	"fmt"
	"testing"
	"time"

	"github.com/Zensey/crawler/pkg/jobs"
	"github.com/Zensey/crawler/pkg/util"
)

func Test_Enrich(t *testing.T) {

	ctx, _ := util.ShutdownCtx()
	rc := NewRouterContext(ctx)

	rc.Route("main").
		Source(func(send func(Exchange)) {
			for {
				select {
				case <-ctx.Done():
					return

				case <-time.After(1 * time.Second):
					fmt.Println("source tick>")
					send(Exchange{Msg: &jobs.Probe{
						"a",
					}})
				}
			}
		}).
		To("enrich-rt").
		Sink(func(j_ Exchange) error {
			fmt.Printf("Sink >> %+v \n", j_)
			return nil
		})

	rc.Route("enrich-rt").
		Process(1)

	rc.Run()

}
