package mophAlertUtils

type AlertResponse struct {
	Message     string `json:"message"`
	MessageCode int    `json:"message_code"`
}

type FlexAlertPayload struct {
	CID      []string      `json:"cid"`
	Messages []FlexMessage `json:"messages"`
}

type FlexMessage struct {
	Type     string      `json:"type"`
	AltText  string      `json:"altText"`
	Contents interface{} `json:"contents"`
}

type FlexBubble struct {
	Type   string   `json:"type"`
	Header *FlexBox `json:"header,omitempty"`
	Body   *FlexBox `json:"body,omitempty"`
	Footer *FlexBox `json:"footer,omitempty"`
}

type FlexBox struct {
	Type     string        `json:"type"`
	Layout   string        `json:"layout"`
	Contents []interface{} `json:"contents"`
	Spacing  string        `json:"spacing,omitempty"`
	Margin   string        `json:"margin,omitempty"`
}

type FlexText struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size   string `json:"size,omitempty"`
	Color  string `json:"color,omitempty"`
	Wrap   bool   `json:"wrap,omitempty"`
	Margin string `json:"margin,omitempty"`
	Flex   int    `json:"flex,omitempty"`
}
