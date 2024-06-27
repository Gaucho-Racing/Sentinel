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
func GetDriveMemberPermission(driveID string, email string) (*drive.Permission, error) {
	resp, err := DriveClient.Permissions.List(driveID).
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
		resp, err = DriveClient.Permissions.List(driveID).
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

// RemoveMemberFromDrive removes a user from the shared drive.
func RemoveMemberFromDrive(driveID string, email string) error {
	perm, err := GetDriveMemberPermission(driveID, email)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	} else if perm == nil {
		return fmt.Errorf("user not found in drive")
	}
	err = DriveClient.Permissions.Delete(driveID, perm.Id).SupportsAllDrives(true).Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	return nil
}

// AddMemberToDrive adds a user to the shared drive with the specified role.
// The role can be "organizer", "fileOrganizer", "writer", "commenter", or "reader".
func AddMemberToDrive(driveID string, email string, role string) error {
	perm, err := GetDriveMemberPermission(driveID, email)
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	} else if perm != nil {
		return fmt.Errorf("user already in drive")
	}
	perm = &drive.Permission{
		EmailAddress: email,
		Role:         role,
		Type:         "user",
	}
	resp, err := DriveClient.Permissions.Create(driveID, perm).SupportsAllDrives(true).Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return err
	}
	utils.SugarLogger.Infof("Permission ID: %s", resp.Id)
	return nil
}
