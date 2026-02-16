package api

type ChunkPayload struct {
	ProjectID        string   `json:"project_id"`
	DocID            string   `json:"doc_id"`
	DocVersion       string   `json:"doc_version"`
	DocVersionTS     int64    `json:"doc_version_ts"`
	IsActive         bool     `json:"is_active"`
	ChunkID          int      `json:"chunk_id"`
	Source           string   `json:"source"`
	Title            string   `json:"title"`
	PathOrURL        string   `json:"path_or_url"`
	Text             string   `json:"text"`
	ContentHash      string   `json:"content_hash"`
	ACLPublic        bool     `json:"acl_public"`
	ACLExternalPublic bool    `json:"acl_external_public"`
	ACLAllow         []string `json:"acl_allow"`
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
	Deleted          bool     `json:"deleted"`
}
