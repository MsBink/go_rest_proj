package user

type Service struct {
	storage Storage
}

//func (s *Service) Register(ctx context.Context, dto CreateUserDTO) (u User, e error) {
//	result := s.storage.FindOne(ctx, dto.Username)
//	if result.Err() == nil {
//		return "", fmt.Errorf("username already exists")
//	} else if !errors.Is(result.Err(), mongo.ErrNoDocuments) {
//		return "", fmt.Errorf("failed to check username availability: %v", result.Err())
//	}
//
//	newUser := User{
//		Username:     credentials.Username,
//		PasswordHash: credentials.Password,
//	}
//
//	id, err := d.Create(ctx, newUser)
//	if err != nil {
//		return "", fmt.Errorf("failed to create user during registration: %v", err)
//	}
//
//	return id, nil
//}
//func (d *db) Register(ctx context.Context, credentials UserCredentials) (string, error) {
//
//}
