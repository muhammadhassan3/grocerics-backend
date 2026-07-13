package dto

// @Swagger:model SignedURL
// @Property presigned_url: Signed URL for uploading the file
// @Property key: Key of the file in the storage bucket
// @Property expires_at: Expiration timestamp of the signed URL, RFC3339
// @Property public_url: Public URL of the uploaded file
// @Description Signed URL for uploading a file to the storage bucket.
type PresignedURL struct {
	// Signed URL for uploading the file
	PresignedURL string `json:"presigned_url"`
	// Key of the file in the storage bucket
	Key string `json:"key"`
	// Expiration timestamp of the signed URL, RFC3339
	ExpiresAt string `json:"expires_at"`
	// Public URL of the uploaded file
	PublicURL string `json:"public_url"`
}
