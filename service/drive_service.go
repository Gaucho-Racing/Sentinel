package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"sentinel/config"
	"sentinel/utils"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var DriveClient *drive.Service

func InitializeDrive() {
	ctx := context.Background()
	decoded, err := base64.StdEncoding.DecodeString(config.DriveServiceAccount)
	if err != nil {
		utils.SugarLogger.Fatalln("Error decoding service account: %v\n", err)
	}
	creds, err := google.CredentialsFromJSON(ctx, []byte(decoded), drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Unable to create Drive service: %v", err)
	}
	DriveClient = srv

	resp, err := srv.Drives.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve shared drives: %v", err)
	}
	if len(resp.Drives) == 0 {
		fmt.Println("No drives found.")
	} else {
		for _, d := range resp.Drives {
			fmt.Printf("%s (%s)\n", d.Name, d.Id)
		}
	}

	// List files in the shared drive
	fresp, err := srv.Files.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	if len(fresp.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		fmt.Println("Files:")
		for _, f := range fresp.Files {
			fmt.Printf("%s (%s)\n", f.Name, f.Id)
		}
	}
}

func GetDriveMemberStatus(email string) (bool, error) {
	// List permissions for the shared drive
	resp, err := DriveClient.Permissions.List("1ao1ErhgJ3-2YcdCheBdOqaltO9X-7QFSnKOiImYC6Hg").
		Fields("permissions(id, type, emailAddress, role)").
		Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return false, err
	}
	println("checking for email: ", email)
	// Check if the email exists in the list of permissions
	for _, perm := range resp.Permissions {
		if perm.EmailAddress == email {
			fmt.Printf("%v", perm)
			return true, nil
		}
	}
	// If email not found in permissions list
	return false, nil
}
