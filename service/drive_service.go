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
}

// GetDriveMemberPermission returns the permissions of the user in the shared drive.
// The included fields are nextPageToken, permissions(id, type, emailAddress, role).
func GetDriveMemberPermission(email string) (*drive.Permission, error) {
	resp, err := DriveClient.Permissions.List(config.SharedDriveID).
		SupportsAllDrives(true).
		Fields("nextPageToken,permissions(id, type, emailAddress, role)").
		Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return nil, err
	}
	for _, perm := range resp.Permissions {
		if perm.EmailAddress == email {
			return perm, nil
		}
	}
	nextPageToken := resp.NextPageToken
	for nextPageToken != "" {
		resp, err = DriveClient.Permissions.List(config.SharedDriveID).
			SupportsAllDrives(true).
			Fields("nextPageToken,permissions(id, type, emailAddress, role)").
			PageToken(nextPageToken).
			Do()
		if err != nil {
			utils.SugarLogger.Errorln(err)
			return nil, err
		}
		for _, perm := range resp.Permissions {
			if perm.EmailAddress == email {
				return perm, nil
			}
		}
		nextPageToken = resp.NextPageToken
	}
	return nil, nil
}

func RemoveMemberFromDrive(email string) error {
	perm, err := GetDriveMemberPermission(email)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	} else if perm == nil {
		return fmt.Errorf("user not found in drive")
	}
	err = DriveClient.Permissions.Delete(config.SharedDriveID, perm.Id).SupportsAllDrives(true).Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	return nil
}

func TestDrive() {
	perm, err := GetDriveMemberPermission("bkathi@ucsb.edu")
	if err != nil {
		utils.SugarLogger.Errorln(err)
	}
	println(perm.Role)

	err = RemoveMemberFromDrive("bkathi@ucsb.edu")
	if err != nil {
		utils.SugarLogger.Errorln(err)
	}
}
