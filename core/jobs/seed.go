package jobs

import (
	"time"

	"github.com/gaucho-racing/sentinel/core/config"
	"github.com/gaucho-racing/sentinel/core/model"
	"github.com/gaucho-racing/sentinel/core/pkg/logger"
	"github.com/gaucho-racing/sentinel/core/service"
	"gorm.io/gorm"
)

// DEV-ONLY seed data. Skipped in production. Designed to be re-runnable —
// every row is checked by fixed ID before insert.
const TestGroupID = "grp_mock00000000000000000seed"

type mockUserSeed struct {
	EntityID       string
	UserID         string
	Username       string
	FirstName      string
	LastName       string
	GraduateLevel  string
	GraduationYear int
	Major          string
	InitialRole    string
}

var mockUsers = []mockUserSeed{
	{
		EntityID: "ent_mock00000000000000000001", UserID: "usr_mock00000000000000000001",
		Username: "achen", FirstName: "Alex", LastName: "Chen",
		GraduateLevel: "Undergraduate", GraduationYear: 2027, Major: "Mechanical Engineering",
		InitialRole: "member",
	},
	{
		EntityID: "ent_mock00000000000000000002", UserID: "usr_mock00000000000000000002",
		Username: "ppatel", FirstName: "Priya", LastName: "Patel",
		GraduateLevel: "Undergraduate", GraduationYear: 2026, Major: "Computer Engineering",
		InitialRole: "member",
	},
	{
		EntityID: "ent_mock00000000000000000003", UserID: "usr_mock00000000000000000003",
		Username: "mjohnson", FirstName: "Marcus", LastName: "Johnson",
		GraduateLevel: "Undergraduate", GraduationYear: 2028, Major: "Electrical Engineering",
		InitialRole: "member",
	},
	{
		EntityID: "ent_mock00000000000000000004", UserID: "usr_mock00000000000000000004",
		Username: "srodriguez", FirstName: "Sofia", LastName: "Rodriguez",
		GraduateLevel: "Undergraduate", GraduationYear: 2027, Major: "Materials Science",
		InitialRole: "member",
	},
	{
		EntityID: "ent_mock00000000000000000005", UserID: "usr_mock00000000000000000005",
		Username: "lkim", FirstName: "Liam", LastName: "Kim",
		GraduateLevel: "Graduate", GraduationYear: 2026, Major: "Aerospace Engineering",
		InitialRole: "member",
	},
}

type joinRequestSeed struct {
	RequestID string
	EntityIdx int // index into mockUsers
	Reason    string
	Duration  time.Duration // how long of access the requester wants
}

var seedJoinRequests = []joinRequestSeed{
	{RequestID: "gjr_mock00000000000000000001", EntityIdx: 1, Reason: "Joining the suspension subteam this quarter — VP Wilson asked me to request access.", Duration: 90 * 24 * time.Hour},
	{RequestID: "gjr_mock00000000000000000002", EntityIdx: 2, Reason: "Working on the damper test rig and need access to the team Drive and Trackside.", Duration: 14 * 24 * time.Hour},
	{RequestID: "gjr_mock00000000000000000003", EntityIdx: 3, Reason: "", Duration: 365 * 24 * time.Hour},
}

func SeedDevData() {
	if config.IsProduction() {
		return
	}
	logger.SugarLogger.Infoln("Seeding dev data (mock users + test group)")
	seedMockUsers()
	seedTestGroup()
	seedTestJoinRequests()
}

func seedMockUsers() {
	for _, mu := range mockUsers {
		if _, err := service.GetEntityByID(mu.EntityID); err == nil {
			continue
		} else if err != gorm.ErrRecordNotFound {
			logger.SugarLogger.Errorf("Failed to check mock entity %s: %v", mu.EntityID, err)
			continue
		}

		if _, err := service.CreateEntity(model.Entity{
			ID:   mu.EntityID,
			Type: model.EntityTypeUser,
		}); err != nil {
			logger.SugarLogger.Errorf("Failed to create mock entity %s: %v", mu.EntityID, err)
			continue
		}

		if _, err := service.CreateUser(model.User{
			ID:             mu.UserID,
			EntityID:       mu.EntityID,
			Username:       mu.Username,
			FirstName:      mu.FirstName,
			LastName:       mu.LastName,
			Gender:         "Unspecified",
			Birthday:       time.Date(2003, 1, 1, 0, 0, 0, 0, time.UTC),
			GraduateLevel:  mu.GraduateLevel,
			GraduationYear: mu.GraduationYear,
			Major:          mu.Major,
			ShirtSize:      "M",
			JacketSize:     "M",
			InitialRole:    mu.InitialRole,
		}); err != nil {
			logger.SugarLogger.Errorf("Failed to create mock user %s: %v", mu.UserID, err)
			continue
		}
		logger.SugarLogger.Infof("Seeded mock user %s (%s %s)", mu.UserID, mu.FirstName, mu.LastName)
	}
}

func seedTestGroup() {
	owner := mockUsers[0]

	_, err := service.GetGroupByID(TestGroupID)
	if err == gorm.ErrRecordNotFound {
		if _, err := service.CreateGroup(model.Group{
			ID:             TestGroupID,
			Name:           "Suspension Team",
			Description:    "Test group for review queue UX — geometry, dampers, kinematics.",
			AllowedSources: model.StringSlice{"DIRECT", "DISCORD"},
			CreatedBy:      owner.EntityID,
		}); err != nil {
			logger.SugarLogger.Errorf("Failed to create test group: %v", err)
			return
		}
		if _, err := service.CreateGroupOwner(model.GroupOwner{
			GroupID:  TestGroupID,
			EntityID: owner.EntityID,
			AddedBy:  owner.EntityID,
		}); err != nil {
			logger.SugarLogger.Errorf("Failed to add owner to test group: %v", err)
		}
		logger.SugarLogger.Infof("Seeded test group %s (owner=%s)", TestGroupID, owner.EntityID)
	} else if err != nil {
		logger.SugarLogger.Errorf("Failed to check test group: %v", err)
		return
	}
}

func seedTestJoinRequests() {
	existing, err := service.GetJoinRequestsByGroup(TestGroupID)
	if err != nil {
		logger.SugarLogger.Errorf("Failed to check existing join requests: %v", err)
		return
	}
	if len(existing) > 0 {
		return
	}

	for _, jr := range seedJoinRequests {
		user := mockUsers[jr.EntityIdx]
		req, err := service.CreateJoinRequest(model.GroupJoinRequest{
			ID:            jr.RequestID,
			GroupID:       TestGroupID,
			EntityID:      user.EntityID,
			Status:        string(model.GroupJoinRequestStatusPending),
			HasExpiration: true,
			ExpiresAt:     time.Now().Add(jr.Duration),
		})
		if err != nil {
			logger.SugarLogger.Errorf("Failed to create mock join request %s: %v", jr.RequestID, err)
			continue
		}
		if jr.Reason != "" {
			if _, err := service.CreateJoinRequestComment(model.GroupJoinRequestComment{
				RequestID: req.ID,
				EntityID:  user.EntityID,
				Comment:   jr.Reason,
			}); err != nil {
				logger.SugarLogger.Errorf("Failed to create reason comment for %s: %v", req.ID, err)
			}
		}
		logger.SugarLogger.Infof("Seeded join request %s from %s", req.ID, user.Username)
	}
}
