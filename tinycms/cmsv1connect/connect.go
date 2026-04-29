package cmsv1connect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"
)

const CMSServiceName = "cms.v1.CMSService"

const (
	CMSServiceHealthProcedure     = "/cms.v1.CMSService/Health"
	CMSServiceListPagesProcedure  = "/cms.v1.CMSService/ListPages"
	CMSServiceGetPageProcedure    = "/cms.v1.CMSService/GetPage"
	CMSServiceSavePageProcedure   = "/cms.v1.CMSService/SavePage"
	CMSServiceDeletePageProcedure = "/cms.v1.CMSService/DeletePage"
	CMSServiceUploadFileProcedure = "/cms.v1.CMSService/UploadFile"
)

type CMSServiceHandler interface {
	Health(context.Context, *connect.Request[structpb.Struct]) (*connect.Response[structpb.Struct], error)
	ListPages(context.Context, *connect.Request[structpb.Struct]) (*connect.Response[structpb.Struct], error)
	GetPage(context.Context, *connect.Request[structpb.Struct]) (*connect.Response[structpb.Struct], error)
	SavePage(context.Context, *connect.Request[structpb.Struct]) (*connect.Response[structpb.Struct], error)
	DeletePage(context.Context, *connect.Request[structpb.Struct]) (*connect.Response[structpb.Struct], error)
	UploadFile(context.Context, *connect.Request[structpb.Struct]) (*connect.Response[structpb.Struct], error)
}

func NewCMSServiceHandler(svc CMSServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(CMSServiceHealthProcedure, connect.NewUnaryHandler(CMSServiceHealthProcedure, svc.Health, opts...))
	mux.Handle(CMSServiceListPagesProcedure, connect.NewUnaryHandler(CMSServiceListPagesProcedure, svc.ListPages, opts...))
	mux.Handle(CMSServiceGetPageProcedure, connect.NewUnaryHandler(CMSServiceGetPageProcedure, svc.GetPage, opts...))
	mux.Handle(CMSServiceSavePageProcedure, connect.NewUnaryHandler(CMSServiceSavePageProcedure, svc.SavePage, opts...))
	mux.Handle(CMSServiceDeletePageProcedure, connect.NewUnaryHandler(CMSServiceDeletePageProcedure, svc.DeletePage, opts...))
	mux.Handle(CMSServiceUploadFileProcedure, connect.NewUnaryHandler(CMSServiceUploadFileProcedure, svc.UploadFile, opts...))
	return "/cms.v1.CMSService/", mux
}
