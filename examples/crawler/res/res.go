package res

import (
	"net/http"
)

type Resources struct {
	Client *http.Client
	//Db     *data.Db
}
