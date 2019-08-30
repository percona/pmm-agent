package tests

// MongoDBDSN returns the DNS for the MongoDB testing instance
// Connection params must match the values in docker-compose.yml
func MongoDBDSN() string {
	return "mongodb://root:root-password@127.0.0.1:27017/admin"
}
