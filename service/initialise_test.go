package service

import (
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-upload-service/config"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var serverWg = &sync.WaitGroup{}
var err = func() error {
	serverWg.Done()
	return errors.New("dlfdlf")
}

func TestGetHTTPServer(t *testing.T) {
	Convey("Given a service list returns a server", t, func() {
		cfg, _ := config.Get()
		//ctx := context.Background()
		ss := &HTTPServerMock{
			ListenAndServeFunc: err,
		}
		newServiceMock := &InitialiserMock{
			DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) HTTPServer {
				return ss
			},
		}
		r := mux.NewRouter()
		//svcErrors := make(chan error, 1)
		//http is called by the service list in  services
		svcList := NewServiceList(newServiceMock)
		//	Run(ctx, svcList, "", "", "", svcErrors)
		svcList.GetHTTPServer(cfg.BindAddr, r)
		So(len(newServiceMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
		So(newServiceMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, cfg.BindAddr)
		//So(len(ss.ListenAndServeCalls()), ShouldEqual, 1)
		//So(ser.ListenAndServe, ShouldBeError)

	})
}

// // GetHTTPServer creates an http server
// func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
// 	s := e.Init.DoGetHTTPServer(bindAddr, router)
// 	return s
// }
