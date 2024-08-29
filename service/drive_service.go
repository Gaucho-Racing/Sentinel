package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"sentinel/config"
	"sentinel/utils"
	"sort"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var DriveClient *drive.Service
var SheetClient *sheets.Service

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

	srv2, err := sheets.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}
	SheetClient = srv2
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

// PopulateDriveMembers adds all users to the shared drive with the appropriate role.
// Useful for when you accidentally remove everyone from the shared drive lmfao
func PopulateDriveMembers() {
	users := GetAllUsers()
	for _, user := range users {
		if user.IsInnerCircle() {
			AddMemberToDrive(config.SharedDriveID, user.Email, "organizer")
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Adding %s to drive with `organizer` role", user.Email))
		} else {
			AddMemberToDrive(config.SharedDriveID, user.Email, "writer")
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Adding %s to drive with `writer` role", user.Email))
		}
	}
}

// CleanDriveMembers removes users from the shared drive that are not in the member directory.
func CleanDriveMembers() {
	keepEmails := []string{
		"sentinel-drive@sentinel-416604.iam.gserviceaccount.com",
		"ucsantabarbarasae@gmail.com",
	}

	resp, err := DriveClient.Permissions.List(config.SharedDriveID).
		SupportsAllDrives(true).
		Fields("nextPageToken,permissions(id, type, emailAddress, role)").
		Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return
	}
	for _, perm := range resp.Permissions {
		user := GetUserByEmail(perm.EmailAddress)
		if user.ID == "" && !contains(keepEmails, perm.EmailAddress) {
			utils.SugarLogger.Infof("Removing %s from drive", perm.EmailAddress)
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removing %s from drive", perm.EmailAddress))
			RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
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
			return
		}
		for _, perm := range resp.Permissions {
			user := GetUserByEmail(perm.EmailAddress)
			if user.ID == "" && perm.EmailAddress != "sentinel-drive@sentinel-416604.iam.gserviceaccount.com" {
				utils.SugarLogger.Infof("Removing %s from drive", perm.EmailAddress)
				RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
			}
		}
		nextPageToken = resp.NextPageToken
	}
}

func PopulateMemberDirectorySheet() {
	// Delete all rows after 5
	clearRange := "A6:O"
	clearRequest := &sheets.ClearValuesRequest{}
	_, err := SheetClient.Spreadsheets.Values.Clear(config.MemberDirectorySheetID, clearRange, clearRequest).Do()
	if err != nil {
		utils.SugarLogger.Errorf("Unable to clear data from sheet: %v", err)
		return
	}
	utils.SugarLogger.Infoln("Rows after 5 have been deleted successfully.")

	users := GetAllUsers()
	sort.Slice(users, func(i, j int) bool {
		return users[i].FirstName < users[j].FirstName
	})

	values := make([][]interface{}, len(users))
	for i, user := range users {
		subteams := make([]string, len(user.Subteams))
		for j, subteam := range user.Subteams {
			subteams[j] = subteam.Name
		}
		subteamString := strings.Join(subteams, ", ")
		roleString := strings.Join(user.Roles, ", ")
		values[i] = []interface{}{
			user.ID,
			user.FirstName,
			user.LastName,
			user.Email,
			user.PhoneNumber,
			user.Gender,
			user.Birthday,
			user.GraduateLevel,
			user.GraduationYear,
			user.Major,
			user.ShirtSize,
			user.JacketSize,
			user.SAERegistrationNumber,
			subteamString,
			roleString,
		}
	}

	writeRange := "A6:O"
	writeRequest := &sheets.ValueRange{
		Values: values,
	}
	_, err = SheetClient.Spreadsheets.Values.Update(config.MemberDirectorySheetID, writeRange, writeRequest).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		utils.SugarLogger.Errorf("Unable to write data to sheet: %v", err)
		return
	}
}
