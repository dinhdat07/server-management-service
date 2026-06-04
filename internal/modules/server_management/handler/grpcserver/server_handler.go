package grpcserver

import (
	"context"
	"errors"
	"time"

	server_managementv1 "server-management-service/gen/go/server_management/v1"
	"server-management-service/internal/modules/server_management/domain"
	"server-management-service/internal/modules/server_management/service"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

type ServerManagementServer struct {
	server_managementv1.UnimplementedServerManagementServiceServer
	serverService service.ServerService
}

func NewServerManagementServer(serverService service.ServerService) *ServerManagementServer {
	return &ServerManagementServer{
		serverService: serverService,
	}
}

func mapError(err error) error {
	if errors.Is(err, service.ErrServerNotFound) {
		return gstatus.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, service.ErrIPv4Exists) || errors.Is(err, service.ErrNameExists) {
		return gstatus.Error(codes.AlreadyExists, err.Error())
	}
	return gstatus.Error(codes.Internal, err.Error())
}

func mapServerToPB(server *domain.Server) *server_managementv1.Server {
	if server == nil {
		return nil
	}
	return &server_managementv1.Server{
		ServerId:      server.ServerID,
		ServerName:    server.ServerName,
		Ipv4:          server.IPv4,
		CurrentStatus: string(server.CurrentStatus),
		CreatedAt:     server.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     server.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *ServerManagementServer) CreateServer(ctx context.Context, req *server_managementv1.CreateServerRequest) (*server_managementv1.CreateServerResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	input := service.CreateServerInput{
		ServerName: req.GetServerName(),
		IPv4:       req.GetIpv4(),
	}

	server, err := s.serverService.CreateServer(ctx, input)
	if err != nil {
		return nil, mapError(err)
	}

	return &server_managementv1.CreateServerResponse{
		Server: mapServerToPB(server),
	}, nil
}

func (s *ServerManagementServer) UpdateServer(ctx context.Context, req *server_managementv1.UpdateServerRequest) (*server_managementv1.UpdateServerResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	input := service.UpdateServerInput{
		ServerName: req.GetServerName(),
		IPv4:       req.GetIpv4(),
	}

	server, err := s.serverService.UpdateServer(ctx, req.GetServerId(), input)
	if err != nil {
		return nil, mapError(err)
	}

	return &server_managementv1.UpdateServerResponse{
		Server: mapServerToPB(server),
	}, nil
}

func (s *ServerManagementServer) DeleteServer(ctx context.Context, req *server_managementv1.DeleteServerRequest) (*server_managementv1.DeleteServerResponse, error) {
	if req == nil {
		return nil, gstatus.Error(codes.InvalidArgument, "request is required")
	}

	err := s.serverService.DeleteServer(ctx, req.GetServerId())
	if err != nil {
		return nil, mapError(err)
	}

	return &server_managementv1.DeleteServerResponse{
		Success: true,
	}, nil
}
