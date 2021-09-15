package bootstrap

import (
	"github.com/gorilla/mux"
	"goblog/pck/route"
	"goblog/pck/routes"
)

func SetupRoute() *mux.Router  {
	router := mux.NewRouter()
	routes.RegisterWebRoutes(router)

	route.SetRoute(router)

	return router
}
