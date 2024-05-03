package main

import (
	"context"
	"log"
	"time"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client *permify.Client
var schemaVersion string
var snapToken string

var users = map[string]string{"user1": "password1", "user2": "password2", "user3": "password3"}

func setupPermifyClient() {
	var err error
	client, err = permify.NewClient(
		permify.Config{
			Endpoint: "localhost:3478",
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to create Permify client: %v", err)
	}

	initPermifySchema()
}

func initPermifySchema() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	schema := `
    entity user {}
    entity organization {
        relation group @group
        relation document @document
        relation administrator @user @group#direct_member @group#manager
        relation direct_member @user
        permission admin = administrator
        permission member = direct_member or administrator or group.member
    }
    entity group {
        relation manager @user @group#direct_member @group#manager
        relation direct_member @user @group#direct_member @group#manager
        permission member = direct_member or manager
    }
    entity document {
        relation org @organization
        relation viewer @user @group#direct_member @group#manager
        relation manager @user @group#direct_member @group#manager
        action edit = manager or org.admin
        action view = viewer or manager or org.admin
    }`

	// Writing the schema to the Permify client
	sr, err := client.Schema.Write(ctx, &v1.SchemaWriteRequest{
		TenantId: "t1",
		Schema:   schema,
	})
	if err != nil {
		log.Fatalf("Failed to write schema: %v", err)
	}
	schemaVersion = sr.SchemaVersion
	log.Printf("Schema version %s written successfully", schemaVersion)

	// Setting up relationship tuples
	rr, err := client.Data.Write(ctx, &v1.DataWriteRequest{
		TenantId: "t1",
		Metadata: &v1.DataWriteRequestMetadata{
			SchemaVersion: schemaVersion,
		},
		Tuples: []*v1.Tuple{
			{
				Entity:   &v1.Entity{Type: "document", Id: "1"},
				Relation: "viewer",
				Subject:  &v1.Subject{Type: "user", Id: "user1"},
			},
			{
				Entity:   &v1.Entity{Type: "document", Id: "1"},
				Relation: "manager",
				Subject:  &v1.Subject{Type: "user", Id: "user3"},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to write data tuples: %v", err)
	}
	snapToken = rr.SnapToken
	log.Printf("Data tuples written successfully, snapshot token: %s", snapToken)
}

func checkPermission(username, permission string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkResult, err := client.Permission.Check(ctx, &v1.PermissionCheckRequest{
		TenantId: "t1",
		Entity: &v1.Entity{
			Type: "document",
			Id:   "1",
		},
		Permission: permission,
		Subject: &v1.Subject{
			Type: "user",
			Id:   username,
		},
		Metadata: &v1.PermissionCheckRequestMetadata{
			SnapToken:     snapToken,
			SchemaVersion: schemaVersion,
			Depth:         50,
		},
	})
	if err != nil {
		log.Printf("Failed to check permission '%s' for user '%s': %v", permission, username, err)
		return false
	}
	return checkResult.Can == v1.CheckResult_CHECK_RESULT_ALLOWED
}
