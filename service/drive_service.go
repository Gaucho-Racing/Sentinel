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
	"time"

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
		} else if user.IsMember() {
			AddMemberToDrive(config.SharedDriveID, user.Email, "writer")
		}
	}
}

// RemoveInactiveMembersFromDrive removes inactive users from the shared drive.
func RemoveInactiveMembersFromDrive() {
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
		} else {
			inactivityThreshold := time.Now().AddDate(0, 0, -180)
			if user.UpdatedAt.After(inactivityThreshold) {
				continue
			}
			lastActivity := GetLastActivityForUser(user.ID)
			if lastActivity.ID != "" && lastActivity.CreatedAt.After(inactivityThreshold) {
				continue
			}
			lastLogins := GetLastNLoginsForUser(user.ID, 1)
			if len(lastLogins) > 0 && lastLogins[0].ID != "" && lastLogins[0].CreatedAt.After(inactivityThreshold) {
				continue
			}
			utils.SugarLogger.Infof("Notifying user %s and removing them from drive due to inactivity.", perm.EmailAddress)
			RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
			SendDirectMessage(user.ID, "You have been automatically removed from our shared Google Drive! Due to Google Drive's member limits, we periodically reset access after 180 days of Sentinel or Discord inactivity. **However, you can easily regain access by using the** `!drive` **command in our #roles channel!**")
			SendMessage(config.DiscordLogChannel, fmt.Sprintf("Sent inactivity google drive removal notice to %s (%s %s)", user.ID, user.FirstName, user.LastName))
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
			} else {
				inactivityThreshold := time.Now().AddDate(0, 0, -180)
				if user.UpdatedAt.After(inactivityThreshold) {
					continue
				}
				lastActivity := GetLastActivityForUser(user.ID)
				if lastActivity.ID != "" && lastActivity.CreatedAt.After(inactivityThreshold) {
					continue
				}
				lastLogins := GetLastNLoginsForUser(user.ID, 1)
				if len(lastLogins) > 0 && lastLogins[0].ID != "" && lastLogins[0].CreatedAt.After(inactivityThreshold) {
					continue
				}
				utils.SugarLogger.Infof("Notifying user %s and removing them from drive due to inactivity.", perm.EmailAddress)
				RemoveMemberFromDrive(config.SharedDriveID, perm.EmailAddress)
				SendDirectMessage(user.ID, "You have been automatically removed from our shared Google Drive! Due to Google Drive's member limits, we periodically reset access after 180 days of Sentinel or Discord inactivity. **However, you can easily regain access by using the** `!drive` **command in our #roles channel!**")
				SendMessage(config.DiscordLogChannel, fmt.Sprintf("Sent inactivity google drive removal notice to %s (%s %s)", user.ID, user.FirstName, user.LastName))
			}
		}
		nextPageToken = resp.NextPageToken
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
		} else if user.IsMember() {
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
			} else if user.IsMember() {
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
		nextPageToken = resp.NextPageToken
	}
}

func PopulateMemberDirectorySheet() {
	// Helper function to clear and populate a sheet
	populateUserSheet := func(sheetName string, users []model.User) {
		// Get sheet ID by name
		spreadsheet, err := SheetClient.Spreadsheets.Get(config.MemberDirectorySheetID).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to get spreadsheet: %v", err)
			return
		}

		sheetId := -1
		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == sheetName {
				utils.SugarLogger.Infof("Found sheet %s: %v", sheet.Properties.Title, sheet.Properties.SheetId)
				sheetId = int(sheet.Properties.SheetId)
				break
			}
		}
		if sheetId == -1 {
			utils.SugarLogger.Errorf("Sheet %s not found", sheetName)
			return
		}

		// Clear existing data using sheet ID
		clearRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					UpdateCells: &sheets.UpdateCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          int64(sheetId),
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

	allUsers := GetAllUsers()

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
}

func PopulateMailingListSheet() {
	// Helper function to clear and populate a sheet

	populateMailingListSheet := func(sheetName string, entries []model.MailingList) {
		// Get sheet ID by name
		spreadsheet, err := SheetClient.Spreadsheets.Get(config.MailingListSheetID).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to get spreadsheet: %v", err)
			return
		}

		sheetId := -1
		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == sheetName {
				utils.SugarLogger.Infof("Found sheet %s: %v", sheet.Properties.Title, sheet.Properties.SheetId)
				sheetId = int(sheet.Properties.SheetId)
				break
			}
		}
		if sheetId == -1 {
			utils.SugarLogger.Errorf("Sheet %s not found", sheetName)
			return
		}

		// Clear existing data using sheet ID
		clearRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					UpdateCells: &sheets.UpdateCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          int64(sheetId),
							StartRowIndex:    5, // A6 starts at index 5
							StartColumnIndex: 0, // A column
							EndColumnIndex:   5, // F column
						},
						Fields: "userEnteredValue",
					},
				},
			},
		}

		_, err = SheetClient.Spreadsheets.BatchUpdate(config.MailingListSheetID, clearRequest).Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to clear data from sheet %s: %v", sheetName, err)
			return
		}

		// Prepare values
		values := make([][]interface{}, len(entries))
		for i, entry := range entries {
			values[i] = []interface{}{
				entry.Email,
				entry.FirstName,
				entry.LastName,
				entry.Role,
				entry.Organization,
			}
		}

		// Write data (can still use A1 notation for updates as it's more convenient)
		writeRange := fmt.Sprintf("'%s'!A6:O", sheetName)
		writeRequest := &sheets.ValueRange{
			Values: values,
		}
		_, err = SheetClient.Spreadsheets.Values.Update(config.MailingListSheetID, writeRange, writeRequest).
			ValueInputOption("RAW").
			Do()
		if err != nil {
			utils.SugarLogger.Errorf("Unable to write data to sheet %s: %v", sheetName, err)
			return
		}

		utils.SugarLogger.Infof("Successfully populated %s sheet with %d emails", sheetName, len(entries))
		SendMessage(config.DiscordLogChannel, fmt.Sprintf("Successfully populated `%s` sheet with %d emails", sheetName, len(entries)))
	}

	allMailingListEntries := GetAllMailingListEntries()
	populateMailingListSheet("All", allMailingListEntries)

	externalMailingListEntries := GetExternalMailingListEntries()
	populateMailingListSheet("External", externalMailingListEntries)
}

// UpdateTeamMembers checks the Team Members google sheet and gives the Team Member Discord role to all users with a TRUE cell
func UpdateTeamMembers() {
	sheetName := "New Members"
	readRange := fmt.Sprintf("'%s'!A:H", sheetName)
	resp, err := SheetClient.Spreadsheets.Values.Get(config.TeamMemberMasterListSheetID, readRange).Do()
	if err != nil {
		utils.SugarLogger.Errorf("Unable to read sheet: %v", err)
		return
	}

	var emails []string
	for i, row := range resp.Values {
		// Skip column names
		if i == 0 {
			continue
		}

		// Skip rows that aren't filled until column H
		if len(row) < 8 {
			continue
		}

		// Check if column H is TRUE
		if hValue, ok := row[7].(string); ok && hValue == "TRUE" {
			if email, ok := row[1].(string); ok && email != "" {
				emails = append(emails, email)
			}
		}
	}
	count := 0
	for _, email := range emails {
		user := GetUserByEmail(email)

		if user.ID == "" {
			continue
		}
		if user.IsAlumni() || user.IsTeamMember() || !user.IsMember() {
			continue
		}
		utils.SugarLogger.Infof("Updating %s with Team Member role", email)
		err := Discord.GuildMemberRoleAdd(config.DiscordGuild, user.ID, config.TeamMemberRoleID)
		if err != nil {
			utils.SugarLogger.Errorln("Error adding role, ", err)
			continue
		}
		count++
	}
	utils.SugarLogger.Infof("Gave %d users the Team Member role", count)
	SendMessage(config.DiscordLogChannel, fmt.Sprintf("Gave %d users the Team Member role", count))
}
