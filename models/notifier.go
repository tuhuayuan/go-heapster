package models

// Notifier 就是自定义的LabelSet
type Notifier struct {
	Type   string `json:"type"`
	Config map[string]interface{}
}

// Notifiers 列表
type Notifiers []Notifier
