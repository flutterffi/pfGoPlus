package todo

import "context"

type HTTPEdge interface {
	List(ctx context.Context) (*ListHTTPResponse, error)
	Create(ctx context.Context, req CreateHTTPRequest) (*CreateHTTPResponse, error)
}

type CreateHTTPRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TodoHTTPItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListHTTPResponse struct {
	Items []TodoHTTPItem `json:"items"`
}

type CreateHTTPResponse struct {
	Item TodoHTTPItem `json:"item"`
}

type HTTPAdapter struct {
	service API
}

func NewHTTPAdapter(service API) *HTTPAdapter {
	return &HTTPAdapter{service: service}
}

func (a *HTTPAdapter) List(ctx context.Context) (*ListHTTPResponse, error) {
	items, err := a.service.List(ctx)
	if err != nil {
		return nil, err
	}

	result := &ListHTTPResponse{
		Items: make([]TodoHTTPItem, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, mapHTTPItem(item))
	}
	return result, nil
}

func (a *HTTPAdapter) Create(ctx context.Context, req CreateHTTPRequest) (*CreateHTTPResponse, error) {
	item, err := a.service.Create(ctx, CreateRequest{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &CreateHTTPResponse{
		Item: mapHTTPItem(*item),
	}, nil
}
