package slack

// https://api.slack.com/reference/messaging/payload
type Payload struct {
	Text        string `json:"text"`
	Blocks      []any  `json:"blocks,omitempty"`
	Attachments []any  `json:"attachments,omitempty"`
	ThreadTs    string `json:"thread_ts,omitempty"`
	Mrkdwn      bool   `json:"mrkdwn,omitempty"`
}

// https://api.slack.com/reference/block-kit/blocks#context
type ContextBlock struct {
	Type     string `json:"type"`
	Elements []any  `json:"elements"`
	BlockID  string `json:"block_id,omitempty"`
}

// https://api.slack.com/reference/messaging/block-elements#image
type ImageElement struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
	AltText  string `json:"alt_text"`
}

// https://api.slack.com/reference/messaging/composition-objects#text
type TextObject struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Emoji    bool   `json:"emoji,omitempty"`
	Verbatim bool   `json:"verbatim,omitempty"`
}

// https://api.slack.com/reference/block-kit/blocks#section
type SectionBlock struct {
	Type      string       `json:"type"`
	Text      TextObject   `json:"text,omitempty"`
	BlockID   string       `json:"block_id,omitempty"`
	Fields    []TextObject `json:"fields,omitempty"`
	Accessory any          `json:"accessory,omitempty"`
}
