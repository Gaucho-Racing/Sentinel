package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"sentinel/config"
	"sentinel/model"
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
	SendMessage(config.DiscordLogChannel, fmt.Sprintf("Removed %s from drive", email))
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
	SendMessage(config.DiscordLogChannel, fmt.Sprintf("Added %s to drive with `%s` role", email, role))
	return nil
}

// PopulateDriveMembers adds all users to the shared drive and the leads drive with the appropriate role.
// Useful for when you accidentally remove everyone from the shared drive lmfao
func PopulateDriveMembers() {
	users := GetAllUsers()
	for _, user := range users {
		if user.IsInnerCircle() {
			AddMemberToDrive(config.SharedDriveID, user.Email, "organizer")
			AddMemberToDrive(config.LeadsDriveID, user.Email, "organizer")
		} else if user.IsMember() || user.IsAlumni() {
			AddMemberToDrive(config.SharedDriveID, user.Email, "writer")
		}
	}
}

// CleanDriveMembers removes users from the shared drive that are not in the member directory.
func CleanDriveMembers() {
	keepEmails := []string{
		"sentinel-drive@sentinel-416604.iam.gserviceaccount.com",
		"ucsantabarbarasae@gmail.com",
		"team@gauchoracing.com",
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
		if contains(keepEmails, perm.EmailAddress) {
			utils.SugarLogger.Infof("Keeping %s in drive", perm.EmailAddress)
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Keeping %s in drive", perm.EmailAddress))
			continue
		}
		user := GetUserByEmail(perm.EmailAddress)
		if user.ID == "" {
			utils.SugarLogger.Infof("Removing %s from drive", perm.EmailAddress)
			RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
		} else if user.IsInnerCircle() {
			if perm.Role != "organizer" {
				// User needs organizer role but doesn't currently have it
				utils.SugarLogger.Infof("Updating %s drive permission to organizer", perm.EmailAddress)
				RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
				AddMemberToDrive(config.SharedDriveID, perm.EmailAddress, "organizer")
			}
		} else if user.IsMember() || user.IsAlumni() {
			if perm.Role != "writer" {
				// User needs writer role but doesn't currently have it
				utils.SugarLogger.Infof("Updating %s drive permission to writer", perm.EmailAddress)
				RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
				AddMemberToDrive(config.SharedDriveID, perm.EmailAddress, "writer")
			}
		} else {
			// User is not a member, remove from drive
			utils.SugarLogger.Infof("Removing %s from drive", perm.EmailAddress)
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
			if contains(keepEmails, perm.EmailAddress) {
				utils.SugarLogger.Infof("Keeping %s in drive", perm.EmailAddress)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Keeping %s in drive", perm.EmailAddress))
				continue
			}
			user := GetUserByEmail(perm.EmailAddress)
			if user.ID == "" {
				utils.SugarLogger.Infof("Removing %s from drive", perm.EmailAddress)
				RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
			} else if user.IsInnerCircle() {
				if perm.Role != "organizer" {
					// User needs organizer role but doesn't currently have it
					utils.SugarLogger.Infof("Updating %s drive permission to organizer", perm.EmailAddress)
					RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
					AddMemberToDrive(config.SharedDriveID, perm.EmailAddress, "organizer")
				}
			} else if user.IsMember() || user.IsAlumni() {
				if perm.Role != "writer" {
					// User needs writer role but doesn't currently have it
					utils.SugarLogger.Infof("Updating %s drive permission to writer", perm.EmailAddress)
					RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
					AddMemberToDrive(config.SharedDriveID, perm.EmailAddress, "writer")
				}
			} else {
				// User is not a member, remove from drive
				utils.SugarLogger.Infof("Removing %s from drive", perm.EmailAddress)
				RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
			}
		}
		nextPageToken = resp.NextPageToken
	}
}

// CleanLeadsDriveMembers removes users from the leads drive that are not in the member directory.
func CleanLeadsDriveMembers() {
	keepEmails := []string{
		"sentinel-drive@sentinel-416604.iam.gserviceaccount.com",
		"ucsantabarbarasae@gmail.com",
		"team@gauchoracing.com",
	}

	resp, err := DriveClient.Permissions.List(config.LeadsDriveID).
		SupportsAllDrives(true).
		Fields("nextPageToken,permissions(id, type, emailAddress, role)").
		Do()
	if err != nil {
		utils.SugarLogger.Errorln(err)
		return
	}
	for _, perm := range resp.Permissions {
		if contains(keepEmails, perm.EmailAddress) {
			utils.SugarLogger.Infof("Keeping %s in leads drive", perm.EmailAddress)
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Keeping %s in leads drive", perm.EmailAddress))
			continue
		}
		user := GetUserByEmail(perm.EmailAddress)
		if user.ID == "" {
			utils.SugarLogger.Infof("Removing %s from leads drive", perm.EmailAddress)
			RemoveMemberFromDrive(config.LeadsDriveID, perm.EmailAddress)
		} else if user.IsInnerCircle() {
			if perm.Role != "organizer" {
				// User needs organizer role but doesn't currently have it
				utils.SugarLogger.Infof("Updating %s leads drive permission to organizer", perm.EmailAddress)
				RemoveMemberFromDrive(config.LeadsDriveID, perm.EmailAddress)
				AddMemberToDrive(config.LeadsDriveID, perm.EmailAddress, "organizer")
			}
		} else {
			// User is not inner circle, remove from leads drive
			utils.SugarLogger.Infof("Removing %s from leads drive", perm.EmailAddress)
			RemoveMemberFromDrive(config.LeadsDriveID, perm.EmailAddress)
		}
	}
	nextPageToken := resp.NextPageToken
	for nextPageToken != "" {
		resp, err = DriveClient.Permissions.List(config.LeadsDriveID).
			SupportsAllDrives(true).
			Fields("nextPageToken,permissions(id, type, emailAddress, role)").
			PageToken(nextPageToken).
			Do()
		if err != nil {
			utils.SugarLogger.Errorln(err)
			return
		}
		for _, perm := range resp.Permissions {
			if contains(keepEmails, perm.EmailAddress) {
				utils.SugarLogger.Infof("Keeping %s in drive", perm.EmailAddress)
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Keeping %s in drive", perm.EmailAddress))
				continue
			}
			user := GetUserByEmail(perm.EmailAddress)
			if user.ID == "" {
				utils.SugarLogger.Infof("Removing %s from leads drive", perm.EmailAddress)
				RemoveMemberFromDrive(config.LeadsDriveID, perm.EmailAddress)
			} else if user.IsInnerCircle() {
				if perm.Role != "organizer" {
					// User needs organizer role but doesn't currently have it
					utils.SugarLogger.Infof("Updating %s leads drive permission to organizer", perm.EmailAddress)
					RemoveMemberFromDrive(config.LeadsDriveID, perm.EmailAddress)
					AddMemberToDrive(config.LeadsDriveID, perm.EmailAddress, "organizer")
				}
			} else {
				// User is not inner circle, remove from leads drive
				utils.SugarLogger.Infof("Removing %s from leads drive", perm.EmailAddress)
				RemoveMemberFromDrive(config.LeadsDriveID, perm.EmailAddress)
			}
		}
		nextPageToken = resp.NextPageToken
	}
}

func PopulateMemberDirectorySheet() {
	// Helper functions to clear and populate a sheet
	populateUserSheet := func(sheetName string, users []model.User) {
		// Get sheet ID by name
		spreadsheet, err := SheetClient.Spreadsheets.Get(config.MemberDirectorySheetID).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to get spreadsheet: %v", err)
			return
		}

		var sheetId int64
		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == sheetName {
				sheetId = sheet.Properties.SheetId
				break
			}
		}
		if sheetId == 0 {
			utils.SugarLogger.Errorf("Sheet %s not found", sheetName)
			return
		}

		// Clear existing data using sheet ID
		clearRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					UpdateCells: &sheets.UpdateCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          sheetId,
							StartRowIndex:    5,  // A6 starts at index 5
							StartColumnIndex: 0,  // A column
							EndColumnIndex:   15, // O column
						},
						Fields: "userEnteredValue",
					},
				},
			},
		}

		_, err = SheetClient.Spreadsheets.BatchUpdate(config.MemberDirectorySheetID, clearRequest).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to clear data from sheet %s: %v", sheetName, err)
			return
		}

		// Sort users by first name
		sort.Slice(users, func(i, j int) bool {
			return users[i].FirstName < users[j].FirstName
		})

		// Prepare values
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

		// Write data (can still use A1 notation for updates as it's more convenient)
		writeRange := fmt.Sprintf("'%s'!A6:O", sheetName)
		writeRequest := &sheets.ValueRange{
			Values: values,
		}
		_, err = SheetClient.Spreadsheets.Values.Update(config.MemberDirectorySheetID, writeRange, writeRequest).
			ValueInputOption("RAW").
			Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to write data to sheet %s: %v", sheetName, err)
			return
		}

		utils.SugarLogger.Infof("Successfully populated %s sheet with %d users", sheetName, len(users))
		SendMessage(config.DiscordLogChannel, fmt.Sprintf("Successfully populated `%s` sheet with %d users", sheetName, len(users)))
	}

	populateMailingListSheet := func(sheetName string, entries []model.MailingList) {
		// Get sheet ID by name
		spreadsheet, err := SheetClient.Spreadsheets.Get(config.MemberDirectorySheetID).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to get spreadsheet: %v", err)
			return
		}

		var sheetId int64
		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == sheetName {
				sheetId = sheet.Properties.SheetId
				break
			}
		}
		if sheetId == 0 {
			utils.SugarLogger.Errorf("Sheet %s not found", sheetName)
			return
		}

		// Clear existing data using sheet ID
		clearRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					UpdateCells: &sheets.UpdateCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          sheetId,
							StartRowIndex:    5, // A6 starts at index 5
							StartColumnIndex: 0, // A column
							EndColumnIndex:   1, // B column
						},
						Fields: "userEnteredValue",
					},
				},
			},
		}

		_, err = SheetClient.Spreadsheets.BatchUpdate(config.MemberDirectorySheetID, clearRequest).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to clear data from sheet %s: %v", sheetName, err)
			return
		}

		// Prepare values
		values := make([][]interface{}, len(entries))
		for i, entry := range entries {
			values[i] = []interface{}{
				entry.Email,
			}
		}

		// Write data (can still use A1 notation for updates as it's more convenient)
		writeRange := fmt.Sprintf("'%s'!A6:O", sheetName)
		writeRequest := &sheets.ValueRange{
			Values: values,
		}
		_, err = SheetClient.Spreadsheets.Values.Update(config.MemberDirectorySheetID, writeRange, writeRequest).
			ValueInputOption("RAW").
			Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to write data to sheet %s: %v", sheetName, err)
			return
		}

		utils.SugarLogger.Infof("Successfully populated %s sheet with %d emails", sheetName, len(entries))
		SendMessage(config.DiscordLogChannel, fmt.Sprintf("Successfully populated `%s` sheet with %d emails", sheetName, len(entries)))
	}

	allUsers := GetAllUsers()
	allMailingListEntries := GetAllMailingListEntries()

	// Filter users for each sheet
	var memberUsers []model.User
	var alumniUsers []model.User
	var leadUsers []model.User
	var specialAdvisorUsers []model.User

	for _, user := range allUsers {
		if user.IsMember() {
			memberUsers = append(memberUsers, user)
		}
		if user.HasRole("d_alumni") {
			alumniUsers = append(alumniUsers, user)
		}
		if user.IsLead() || user.IsOfficer() {
			leadUsers = append(leadUsers, user)
		}
		if user.IsSpecialAdvisor() {
			specialAdvisorUsers = append(specialAdvisorUsers, user)
		}
	}

	// Populate each sheet
	populateUserSheet("All", allUsers)
	populateUserSheet("Members", memberUsers)
	populateUserSheet("Alumni", alumniUsers)
	populateUserSheet("Leads", leadUsers)
	populateUserSheet("Special Advisors", specialAdvisorUsers)
	populateMailingListSheet("Mailing List", allMailingListEntries)
}
