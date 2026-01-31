package handler

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/kkonst40/isso/internal/gen/user"
	"github.com/kkonst40/isso/internal/service"
)

type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserGRPCHandler(userService *service.UserService) *UserGRPCHandler {
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
