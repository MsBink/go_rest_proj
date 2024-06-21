package item

type Item struct {
	ID     string `json:"id" bson:"_id,omitempty"`
	Name   string `json:"name" bson:"name"`
	Price  int    `json:"price" bson:"price"`
	UserID string `json:"user_id" bson:"user_id"`
}

type CreateItemDTO struct {
	Name   string `json:"name"`
	Price  int    `json:"price"`
	UserID string `json:"userID"`
}
