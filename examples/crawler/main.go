package main

import (
	"fmt"
	"net/http"

	routing "github.com/Zensey/go-data-routing"
	"github.com/Zensey/go-data-routing/examples/crawler/jobs"
	"github.com/Zensey/go-data-routing/examples/crawler/res"
	"github.com/Zensey/go-data-routing/examples/crawler/service"
	"github.com/Zensey/go-data-routing/examples/crawler/util"
)

func main() {
	res := res.Resources{
		Client: http.DefaultClient,
	}

	ctx, cancel := util.ShutdownCtx()
	count := 0

	vs := service.NewVisitTrackerService()
	vs.AddUrl("https://google.com/") // seed

	rc := routing.NewRouterContext(ctx)
	rc.Route("main").
		Source(func(n *routing.Node) {
			for {
				select {
				case <-ctx.Done():
					return

				case u := <-vs.DequeUrl():
					fmt.Println("DequeUrl >>", u)
					n.Send(routing.Exchange{Msg: &jobs.HtmlDocChecker{
						Url:       u,
						Resources: &res,
					}})

					if count++; count == 2 {
						cancel()
					}
				}
			}
		}).
		Process(1).
		Filter(func(e routing.Exchange, n *routing.Node) {
			j, ok := e.Msg.(*jobs.HtmlDocChecker)
			if !ok {
				return
			}

			newUrls := vs.FilterNewUrls(j.Hrefs, j.Url)
			for _, v := range newUrls {
				fmt.Println("Filter >", v)
			}

			n.Send(routing.Exchange{Msg: &jobs.HtmlDocChecker{
				Url:       j.Url,
				Hrefs:     newUrls,
				Resources: &res,
			}})
		}).
		Sink(func(e routing.Exchange) error {
			j := e.Msg.(*jobs.HtmlDocChecker)
			fmt.Println("Sink: new urls >", len(j.Hrefs))
			return nil
		})

	rc.Run()
	fmt.Println("Exit >")
}
