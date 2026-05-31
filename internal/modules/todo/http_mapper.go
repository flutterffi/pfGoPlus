package todo

import "time"

func mapHTTPItem(item Todo) TodoHTTPItem {
	return TodoHTTPItem{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Status:      item.Status,
		CreatedAt:   formatHTTPTime(item.CreatedAt),
		UpdatedAt:   formatHTTPTime(item.UpdatedAt),
	}
}

func formatHTTPTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
