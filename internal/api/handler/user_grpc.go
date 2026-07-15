package handler

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/kkonst40/sso-service/internal/gen/user"
)

type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	userService UserService
}

func NewUserGRPCHandler(userService UserService) *UserGRPCHandler {
	return &UserGRPCHandler{userService: userService}
}

func (s *UserGRPCHandler) Exist(ctx context.Context, req *pb.ExistRequest) (*pb.ExistResponse, error) {
	var inputIDs []uuid.UUID
	for _, idStr := range req.Ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		inputIDs = append(inputIDs, id)
	}

	existingIDs, err := s.userService.Exist(ctx, inputIDs)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(existingIDs))
	for i, id := range existingIDs {
		result[i] = id.String()
	}

	return &pb.ExistResponse{ExistingIds: result}, nil
}

func (s *UserGRPCHandler) GetUsersLogins(ctx context.Context, req *pb.GetUsersLoginsRequest) (*pb.GetUsersLoginsResponse, error) {
	var inputIDs []uuid.UUID
	for _, idStr := range req.Ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		inputIDs = append(inputIDs, id)
	}

	usersData, err := s.userService.GetLoginsByIDs(ctx, inputIDs)
	if err != nil {
		return nil, err // обернуть в status.Error, если gRPC ошибка
	}

	pbUsers := make([]*pb.UserInfo, 0, len(usersData))
	for _, u := range usersData {
		pbUsers = append(pbUsers, &pb.UserInfo{
			Id:    u.ID.String(),
			Login: u.Login,
		})
	}

	return &pb.GetUsersLoginsResponse{
		Users: pbUsers,
	}, nil
}

func (s *UserGRPCHandler) GetUsersIDs(ctx context.Context, req *pb.GetUsersIDsRequest) (*pb.GetUsersIDsResponse, error) {
	usersData, err := s.userService.GetIDsByLogins(ctx, req.Logins)
	if err != nil {
		return nil, err // обернуть в status.Error, если gRPC ошибка
	}

	pbUsers := make([]*pb.UserInfo, 0, len(usersData))
	for _, u := range usersData {
		pbUsers = append(pbUsers, &pb.UserInfo{
			Id:    u.ID.String(),
			Login: u.Login,
		})
	}

	return &pb.GetUsersIDsResponse{
		Users: pbUsers,
	}, nil
}
