package storage

const githubWebhookQuery = `
INSERT INTO github (
	delivery_id, 
	event, 
	target, 
	target_id, 
	hook_id, 
	user_login, 
	user_id, 
	"offset", 
	partition, 
	payload
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
